package alg

// 模拟退火算法实现

import (
	"log"
	"math"
	"math/rand"
	"time"
)

var DEBUG = false

// 执行策略
type SAStrategy interface {
	// 通过变异得到一个子样本
	VariantSA() SAStrategy
	// 获取当前策略的得分
	Score() float64
}

// 带有得分缓存的执行策略
type CachedSAStrategy struct {
	ref   SAStrategy
	score float64
}

func (c *CachedSAStrategy) VariantSA() *CachedSAStrategy {
	newRef := c.ref.VariantSA()
	return &CachedSAStrategy{ref: newRef, score: newRef.Score()}
}

func (c *CachedSAStrategy) Score() float64 {
	return c.score
}

// 模拟退火算法实现
func RunSimulatedAnnealing(base SAStrategy, nIter, nSubIter int, T, k float64) (SAStrategy, float64) {
	seed := time.Now().UnixMilli()
	rand.Seed(seed)
	best := &CachedSAStrategy{ref: base, score: base.Score()}
	current := best

	log.Printf("SA Start: seed %d, score %.3f, %s\n", seed, current.score, current.ref)

	for i := 0; i < nIter; i++ {
		for j := 0; j < nSubIter; j++ {
			newStrategy := current.VariantSA()
			delta := newStrategy.score - current.score
			// 得分升高, 接受新解
			if delta > 0 {
				current = newStrategy
				if current.score > best.score {
					if DEBUG {
						log.Printf("Update New Best(%.3f -> %.3f): %s\n", best.score, current.score, current.ref)
					}
					best = current
				}
			} else {
				chance := math.Exp(delta / T)
				if rand.Float64() < chance {
					current = newStrategy
				}
			}
		}
		log.Printf("SA round %2d end, T %.3f, chance(delta=-1) %.4e, score %.3f, best score %.3f\n", i, T, math.Exp(-1/T), current.score, best.score)
		T *= k
	}
	log.Printf("SA end, best score %.3f: %s\n", best.score, best.ref)
	return best.ref, best.score
}
