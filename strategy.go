package neptune

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"neptune/alg"
	"sort"
	"strings"
	"time"
)

// 基本策略: 根据顺序选择使用的目标
type SimpleStrategy struct {
	ctx   *ResearchContext
	Ranks []int
	size  int
}

func NewSimpleStrategy(ctx *ResearchContext, ranks []int) *SimpleStrategy {
	size := len(ranks)
	s := SimpleStrategy{ctx: ctx, Ranks: make([]int, size), size: size}
	copy(s.Ranks, ranks)
	return &s
}

func RanksFromDesc(desc string) []int {
	rankMap := make(map[string]int)
	lines := strings.Split(strings.Trim(desc, " \r\n"), "\n")
	items := strings.Split(lines[len(lines)-1], " ")
	for idx, val := range items {
		rankMap[val] = idx
	}
	ranks := make([]int, len(items))
	for tgt, info := range targets {
		ranks[tgt] = rankMap[info.Code]
	}
	return ranks
}

func (s SimpleStrategy) String() string {
	return "S{" + s.Repr() + "}"
}

func (s SimpleStrategy) Repr() string {
	codes := make([]string, s.size)
	for i, val := range SortedItems(s.Ranks) {
		codes[i] = val.Code()
	}
	return strings.Join(codes, " ")
}

type SortHelper struct {
	orders []TargetType
	ranks  []int
}

func SortedItems(ranks []int) []TargetType {
	size := len(ranks)
	sh := &SortHelper{}
	sh.ranks = make([]int, size)
	copy(sh.ranks, ranks)
	sh.orders = make([]TargetType, size)
	for tgt, rank := range ranks {
		sh.orders[rank] = TargetType(tgt)
	}
	sort.Sort(sh)
	return sh.orders
}

func (s *SortHelper) Len() int {
	return len(s.orders)
}

func (s *SortHelper) Less(i, j int) bool {
	return s.ranks[s.orders[i]] < s.ranks[s.orders[j]]
}
func (s *SortHelper) Swap(i, j int) {
	tmp := s.orders[i]
	s.orders[i] = s.orders[j]
	s.orders[j] = tmp
}

func (s *SimpleStrategy) SwapRank(a, b int) {
	// 交换当前位于 a 和 b 的两个策略
	tmp := s.Ranks[a]
	s.Ranks[a] = s.Ranks[b]
	s.Ranks[b] = tmp
}

// 对一次刷新结果应用该策略, 返回将执行的操作
func (s *SimpleStrategy) Apply(stats *ResearchStatus) TargetType {
	best := TARGET_REFRESH
	bestRank := s.size
	for _, tgt := range stats.CurrentTargets {
		rank := s.Ranks[tgt]
		if rank < bestRank {
			best = tgt
			bestRank = rank
		}
	}
	// 现有最优选择序号大于刷新操作序号, 并且可以刷新, 执行刷新操作
	if bestRank > s.Ranks[TARGET_REFRESH] && stats.CanRefresh() {
		return TARGET_REFRESH
	}
	return best
}

func (s *SimpleStrategy) Variant() *SimpleStrategy {
	// 基于当前策略生成一个变种
	variant := SimpleStrategy{Ranks: make([]int, s.size), size: s.size, ctx: s.ctx}
	copy(variant.Ranks, s.Ranks)
	// 随机选择交换目标
	for i := 0; i < 1; i++ {
		idx := rand.Intn(variant.size)
		idx2 := rand.Intn(variant.size - 1)
		if idx2 >= idx {
			idx2++
		}
		variant.SwapRank(idx, idx2)
	}
	return &variant
}

func (s SimpleStrategy) VariantSA() alg.SAStrategy {
	return s.Variant()
}

// 计算当前策略的得分
func (s SimpleStrategy) Score() float64 {
	return s.ctx.GetScore(&s)
}

// 复合策略: 根据条件从子策略中选择一个并执行
type ChangingStrategy struct {
	ctx      *ResearchContext
	Idx      int
	ChangeAt []int
	Sub      []*SimpleStrategy
}

func NewChangingStrategy(ctx *ResearchContext, ranks1 []int, ranks2 []int) *ChangingStrategy {
	st := ChangingStrategy{ctx: ctx, ChangeAt: []int{18, 19, 20, 21, 22}, Sub: []*SimpleStrategy{
		NewSimpleStrategy(ctx, ranks1),
		NewSimpleStrategy(ctx, ranks2),
	}}
	return &st
}

func ChangingStrategyFromJson(ctx *ResearchContext, data []byte) *ChangingStrategy {
	st := ChangingStrategy{ctx: ctx}
	err := json.Unmarshal(data, &st)
	if err != nil {
		panic(err)
	}
	for _, sub := range st.Sub {
		sub.size = len(sub.Ranks)
		sub.ctx = ctx
	}
	return &st
}

// 对一次刷新结果应用该策略, 返回将执行的操作
func (s *ChangingStrategy) Apply(stats *ResearchStatus) TargetType {
	hour := s.ChangeAt[s.Idx]
	curHour := int(stats.CurrentTime/time.Hour) % 24
	if curHour < hour {
		return s.Sub[0].Apply(stats)
	}
	return s.Sub[1].Apply(stats)
}

func (s *ChangingStrategy) Variant() *ChangingStrategy {
	v := ChangingStrategy{Idx: s.Idx, ChangeAt: s.ChangeAt, Sub: make([]*SimpleStrategy, len(s.Sub)), ctx: s.ctx}
	v.Sub[0] = s.Sub[0]
	v.Sub[1] = s.Sub[1]

	chance := rand.Float64()
	if chance < 0.2 {
		// 修改Idx
		v.Idx = rand.Intn(len(v.ChangeAt))
		return &v
	} else if chance < 0.6 {
		v.Sub[0] = v.Sub[0].Variant()
	} else {
		v.Sub[1] = v.Sub[1].Variant()
	}
	return &v
}

func (s *ChangingStrategy) VariantSA() alg.SAStrategy {
	return s.Variant()
}

func (s *ChangingStrategy) String() string {
	return fmt.Sprintf("CS{@%d %s; %s}", s.ChangeAt[s.Idx], s.Sub[0], s.Sub[1])
}

// func (s *ChangingStrategy) VariantGA() alg.GAStrategy {
// 	return s.Variant()
// }

// 计算当前策略的得分
func (s *ChangingStrategy) Score() float64 {
	return s.ctx.GetScore(s)
}
