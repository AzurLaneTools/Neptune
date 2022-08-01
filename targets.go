package neptune

import (
	"bytes"
	"encoding/csv"
	"log"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "embed"
)

type TargetType int

const TARGET_REFRESH TargetType = 0

type TargetInfo struct {
	Code     string
	Name     string
	Duration time.Duration
	Weight   float64
	CostGain []float64
}

var (
	targets       []TargetInfo
	targetWeights []float64
	TargetHeaders = strings.Split("code,duration,rate,cost-Coins,cost-Cube,gain-BP-SSR,gain-BP-UR,gain-EquipBP-UR,gain-Chips", ",")
)

func (t TargetType) Code() string {
	return targets[t].Code
}

func (t TargetType) String() string {
	info := targets[t]
	return info.Code
}

func (t TargetType) Name() string {
	return targets[t].Name
}

func (t TargetType) Duration() time.Duration {
	return targets[t].Duration
}

func (t TargetType) CostGain() []float64 {
	return targets[t].CostGain
}

func toFloat(text string) float64 {
	text = strings.Trim(text, " ")
	ret, err := strconv.ParseFloat(text, 64)
	if err != nil {
		panic(err)
	}
	return ret
}

func toFloats(items []string) []float64 {
	nums := make([]float64, len(items))
	for i, text := range items {
		nums[i] = toFloat(text)
	}
	return nums
}

//go:embed data.csv
var defaultData []byte

func InitDefaultTargets() {
	targets = make([]TargetInfo, 0)
	csvLines, err := csv.NewReader(bytes.NewBuffer(defaultData)).ReadAll()
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(TargetHeaders, csvLines[0]) {
		log.Println(TargetHeaders, csvLines[0])
		panic("输入数据不符合需求!")
	}
	InitTargets(csvLines[1:])
}

func InitTargets(rows [][]string) {
	nCols := len(TargetHeaders)
	targets = append(targets, TargetInfo{"SKIP", "刷新", 0, 0, make([]float64, nCols-3)})
	for _, line := range rows {
		if len(line) < nCols {
			panic("长度不匹配!")
		}
		targets = append(targets,
			TargetInfo{
				Code:     line[0],
				Duration: time.Duration(toFloat(line[1])) * time.Hour,
				Weight:   toFloat(line[2]),
				CostGain: toFloats(line[3:nCols]),
			})
	}
	targetWeights = make([]float64, len(targets))
	cumsum := float64(0)
	for i := range targets {
		cumsum += targets[i].Weight
		targetWeights[i] = cumsum
	}
	for i := range targets {
		targetWeights[i] = targetWeights[i] / cumsum
	}
	log.Printf("Inited: %v\n", targets)
}

func RandomSample(n int) []TargetType {
	// number of random draws
	var val float64
	indices := make([]TargetType, n)
	// loop through indices and draw random values
	for i := range indices {
		// multiply the sample with the largest CDF value; easier than normalizing to [0,1)
		val = rand.Float64()
		// Search returns the smallest index i such that cdf[i] > val
		indices[i] = TargetType(sort.Search(len(targetWeights), func(i int) bool { return targetWeights[i] > val }))
	}
	return indices
}
