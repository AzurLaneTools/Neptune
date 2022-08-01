package neptune

import (
	"fmt"
	"log"
	"strings"
	"time"
)

var DEBUG = false

// 可操作时间配置
type TimeStrategy interface {
	NextActiveTime(time.Duration) time.Duration
}

type AlwaysOnline struct{}

func (a *AlwaysOnline) NextActiveTime(t time.Duration) time.Duration {
	return t
}

type SkipNight struct{}

func (a *SkipNight) NextActiveTime(t time.Duration) time.Duration {
	hour := t / time.Hour
	hourMod := hour % 24
	if hourMod < 8 {
		// 早上8点前, 延迟到早上8点
		return (hour + (8 - hourMod)) * time.Hour
	}
	if hourMod >= 23 {
		// 23点后, 延迟到第二天早上8点
		return (hour + (24 + 8 - hourMod)) * time.Hour
	}
	return t
}

// 科研策略
type ResearchStrategy interface {
	// 在当前场景下应用该策略, 返回将执行的操作
	Apply(*ResearchStatus) TargetType
}

// 科研配置信息
type ResearchContext struct {
	TargetCount int
	nIter       int
	ts          TimeStrategy
	weights     []float64
	samples     [][]TargetType
}

func NewResearchContext(nIter int, ts TimeStrategy, weights []float64) *ResearchContext {
	rc := &ResearchContext{len(targets), nIter, ts, weights, make([][]TargetType, 0, nIter)}
	for i := 0; i < nIter; i++ {
		rc.samples = append(rc.samples, RandomSample(5))
	}
	return rc
}

// 当前科研信息
type ResearchStatus struct {
	ctx             *ResearchContext
	idx             int
	CountMap        []int
	CurrentTargets  []TargetType
	LastRefreshDate int64
	CurrentTime     time.Duration
	currentDate     int64
	CostGain        []float64
}

var oneDay = time.Hour * 24

func (rs *ResearchStatus) CanRefresh() bool {
	return rs.currentDate > rs.LastRefreshDate
}
func (rs *ResearchStatus) SampleNext() {
	// 从预计算好的结果中获取下一项
	rs.CurrentTargets = rs.ctx.samples[rs.idx]
	rs.idx++
}

// 在当前状态下, 根据指定的优先级和时间策略执行一次操作
func (rs *ResearchStatus) Apply(strategy ResearchStrategy) {
	var endTime time.Duration
	rs.SampleNext()
	tgt := strategy.Apply(rs)
	// log.Printf("Choosed: %s -> %d %s, (can=%v)\n", rs.CurrentTargets, tgt, tgt, rs.CanRefresh())
	rs.CountMap[tgt]++
	if tgt == TARGET_REFRESH {
		rs.LastRefreshDate = rs.currentDate
		endTime = rs.CurrentTime
	} else {
		endTime = rs.CurrentTime + tgt.Duration()
	}
	nextTime := rs.ctx.ts.NextActiveTime(rs.CurrentTime)
	if nextTime < endTime {
		nextTime = endTime
	}
	rs.CurrentTime = nextTime
	rs.currentDate = int64(rs.CurrentTime / oneDay)

	cg := tgt.CostGain()
	if DEBUG {
		log.Printf("Choosed: %s -> %d %s, costGain=%v\n", rs.CurrentTargets, tgt, tgt, cg)
	}
	for i := range rs.CostGain {
		rs.CostGain[i] += cg[i]
	}
}
func (rs ResearchStatus) String() string {
	costMap := make([]string, len(rs.CountMap))
	for i, cnt := range rs.CountMap {
		costMap[i] = fmt.Sprintf("%s=%d", TargetType(i), cnt)
	}
	return fmt.Sprintf("ResearchStatus{N=%d, CountMap={%s}, Duration=%s(%d days), CostGain=%v}", rs.idx, strings.Join(costMap, ","), rs.CurrentTime, rs.currentDate, rs.CostGain)
}

// 在当前状态下, 根据指定的优先级和时间策略执行一次操作
func (rs *ResearchContext) GetStats(strategy ResearchStrategy) ResearchStatus {
	stat := ResearchStatus{ctx: rs, LastRefreshDate: -1, CostGain: make([]float64, len(rs.weights)), CountMap: make([]int, rs.TargetCount)}
	// log.Printf("Start at %s\n", stat.CurrentTime)
	for i := 0; i < rs.nIter; i++ {
		stat.Apply(strategy)
	}
	return stat
}

// 在当前状态下, 根据指定的优先级和时间策略执行一次操作
func (rs *ResearchContext) GetScore(strategy ResearchStrategy) float64 {
	stat := rs.GetStats(strategy)
	// log.Printf("End at %s\n", stat.CurrentTime)
	hours := float64(stat.CurrentTime / time.Hour)

	var score float64
	for i := range rs.weights {
		score += rs.weights[i] * stat.CostGain[i] / hours
	}
	if DEBUG {
		log.Printf("Get Score: %s %.3f in %.3f\n", strategy, score, hours)
	}
	return score
}
