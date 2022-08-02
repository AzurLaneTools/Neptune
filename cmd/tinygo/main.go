package main

import (
	"log"
	"math/rand"
	"neptune"
	"neptune/alg"
	"time"
)

// export function to js runtime.
//export Simulate
func Simulate(data string) string {
	// 投入物资	投入魔方	产出金图	产出彩图	产出彩装备	产出心智单元
	// cost-Coins,cost-Cube,gain-BP-SSR,gain-BP-UR,gain-EquipBP-UR,gain-Chips
	weights := []float64{-0.0001, -3.0, 1.6, 4.0, 150.0, 0.00001}
	// weights := []float64{-0.001, 0, 1, 1, 365 * 4, 0.01}
	ctx := neptune.NewResearchContext(3000, &neptune.AlwaysOnline{}, weights)
	ranks := make([]int, 0)
	for i := 0; i < ctx.TargetCount; i++ {
		ranks = append(ranks, i)
	}
	rand.Shuffle(len(ranks), func(i, j int) {
		ranks[i], ranks[j] = ranks[j], ranks[i]
	})
	base := neptune.NewSimpleStrategy(ctx, ranks)
	best, bestScore := alg.RunSimulatedAnnealing(base, 100, 200, 1.0, 0.94)
	log.Printf("Run End. Best(%.3f): %s\n", bestScore, best)
	// log.Printf("Seed: %d\n", seed)

	st := best.(*neptune.SimpleStrategy)
	stat := ctx.GetStats(st)
	log.Printf("stat: %+v\n", stat)
	nDay := float64(stat.CurrentTime) / float64(time.Hour*24)
	log.Printf("Daily Cost & Gain:\n")
	for i, val := range stat.CostGain {
		log.Printf("%s: %.3f\n", neptune.TargetHeaders[i+3], val/nDay)
	}
	return ""
}

func main() {
	neptune.InitDefaultTargets()
}
