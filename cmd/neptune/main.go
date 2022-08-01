package main

import (
	"encoding/json"
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
	weights := []float64{-0.0001, -0.6 * 3, 1.6, 3.0, 36 * 4, 0.00001}
	// weights := []float64{-0.001, 0, 1, 1, 365 * 4, 0.01}
	ctx := neptune.NewResearchContext(10000, &neptune.SkipNight{}, weights)

	saveName := fmt.Sprintf("records/%s.json", os.Args[1])

	jsonData, err := os.ReadFile(saveName)
	var base *neptune.ChangingStrategy
	if err != nil {
		log.Println(err)
		ranks := make([]int, 0)
		for i := 0; i < ctx.TargetCount; i++ {
			ranks = append(ranks, i)
		}
		rand.Shuffle(len(ranks), func(i, j int) {
			ranks[i], ranks[j] = ranks[j], ranks[i]
		})
		// recData = make([]byte, 0)
		base = neptune.NewChangingStrategy(ctx, ranks, ranks)
	} else {
		base = neptune.ChangingStrategyFromJson(ctx, jsonData)
		log.Printf("Loaded %v\n", base)
		// if recData[len(recData)-1] != '\n' {
		// 	recData = append(recData, '\n')
		// }
	}
	// base := neptune.NewSimpleStrategy(ctx, ranks)
	curScore := base.Score()
	log.Printf("Start: %.3f %v\n", curScore, base)
	for nRun := 0; nRun < 1; nRun++ {
		best, bestScore := alg.RunSimulatedAnnealing(base, 100, 200, 1, 0.94)
		log.Printf("Best(%.3f): %s\n", bestScore, best)

		base = best.(*neptune.ChangingStrategy)
		stat := ctx.GetStats(base)
		log.Printf("stat: %s\n", stat)
		nDay := float64(stat.CurrentTime) / float64(time.Hour*24)
		log.Printf("Daily Cost & Gain:\n")
		for i, val := range stat.CostGain {
			log.Printf("%s: %.3f\n", neptune.TargetHeaders[i+3], val/nDay)
		}
		log.Printf("Score Change: %.3f -> %.3f", curScore, bestScore)
		if curScore < bestScore {
			data, err := json.Marshal(base)
			if err != nil {
				panic(err)
			}
			os.WriteFile(saveName, data, 0755)
			log.Printf("Write %s\n", data)
			curScore = bestScore
		}
	}
}
