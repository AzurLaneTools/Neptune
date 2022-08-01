package main

import (
	"fmt"
	"log"
	"math/rand"
	"neptune"
	"neptune/alg"
	"os"
	"time"
)

func main() {
	log.Println("开始运行")
	neptune.InitDefaultTargets()
	neptune.DEBUG = false
	// cost-Coins,cost-Cube,gain-BP-SSR,gain-BP-UR,gain-EquipBP-UR,gain-CognitiveChips
	weights := []float64{-0.0001, -3.0, 1.6, 4.0, 150.0, 0.00001}
	// weights := []float64{-0.001, 0, 1, 1, 365 * 4, 0.01}
	ctx := neptune.NewResearchContext(5000, &neptune.AlwaysOnline{}, weights)
	// ctx := neptune.NewResearchContext(3000, &neptune.SkipNight{}, weights)

	saveName := fmt.Sprintf("records/%s.txt", os.Args[1])

	var ranks []int
	recData, err := os.ReadFile(saveName)
	if err != nil {
		ranks = make([]int, 0)
		for i := 0; i < ctx.TargetCount; i++ {
			ranks = append(ranks, i)
		}
		rand.Shuffle(len(ranks), func(i, j int) {
			ranks[i], ranks[j] = ranks[j], ranks[i]
		})
		recData = make([]byte, 0)
	} else {
		ranks = neptune.RanksFromDesc(string(recData))

		if recData[len(recData)-1] != '\n' {
			recData = append(recData, '\n')
		}
	}
	// base := neptune.NewChangingStrategy(ctx, ranks, ranks)
	base := neptune.NewSimpleStrategy(ctx, ranks)
	curScore := base.Score()
	for nRun := 0; nRun < 2; nRun++ {
		best, bestScore := alg.RunSimulatedAnnealing(base, 100, 200, 1, 0.94)
		log.Printf("Best(%.3f): %s\n", bestScore, best)

		st := best.(*neptune.SimpleStrategy)
		stat := ctx.GetStats(st)
		log.Printf("stat: %+v\n", stat)
		nDay := float64(stat.CurrentTime) / float64(time.Hour*24)
		log.Printf("Daily Cost & Gain:\n")
		for i, val := range stat.CostGain {
			log.Printf("%s: %.3f\n", neptune.TargetHeaders[i+3], val/nDay)
		}
		log.Printf("Score Change: %.3f -> %.3f", curScore, bestScore)
		if bestScore > curScore {
			recData = append(recData, []byte(best.(*neptune.SimpleStrategy).Repr())...)
			recData = append(recData, '\n')
			os.WriteFile(saveName, recData, 0755)
			log.Printf("Write to %s:\n%s", saveName, recData)
			curScore = bestScore
		}
		recData, err := os.ReadFile(saveName)
		if err != nil {
			panic(err)
		}
		ranks = neptune.RanksFromDesc(string(recData))
		base = neptune.NewSimpleStrategy(ctx, ranks)
	}
}
