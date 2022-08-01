package main

import (
	"log"
	"math/rand"
	"neptune"
	"neptune/alg"
	"syscall/js"
	"time"
)

func jsonWrapper(subfunc func(string) (interface{}, error)) js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid number of arguments passed"
		}
		inputData := args[0].String()
		outputData, err := subfunc(inputData)
		if err != nil {
			return map[string]interface{}{"code": 1, "error": err.Error()}
		}
		return outputData
	})
	return jsonFunc
}

func Simulate(data string) (interface{}, error) {
	neptune.DEBUG = false
	// 投入物资	投入魔方	产出金图	产出彩图	产出彩装备	产出心智单元
	weights := []float64{-0.001, -6, 16, 30, 365 * 4, 0.01}
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
	best, bestScore := alg.RunSimulatedAnnealing(base, 100, 6, 10.0, 0.9)
	log.Printf("Run End. Best(%.3f): %s\n", bestScore, best)
	// log.Printf("Seed: %d\n", seed)

	st := best.(neptune.SimpleStrategy)
	stat := ctx.GetStats(&st)
	log.Printf("stat: %+v\n", stat)
	nDay := float64(stat.CurrentTime) / float64(time.Hour*24)
	log.Printf("Daily Cost & Gain:\n")
	for i, val := range stat.CostGain {
		log.Printf("%s: %.3f\n", neptune.TargetHeaders[i+4], val/nDay)
	}
	return "", nil
}

func main() {
	log.Println("开始运行")
	// js.Global().Set("InitTargets", jsonWrapper(InitTargetsFromJson))
	neptune.InitDefaultTargets()
	js.Global().Set("Simulate", jsonWrapper(Simulate))
	done := make(chan struct{})
	<-done
	log.Println("结束运行")
}
