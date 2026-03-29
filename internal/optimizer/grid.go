package optimizer

import (
	"context"
	"quant-agent/internal/model"
	"sync"
)

// GridOptimizer 网格搜索
type GridOptimizer struct{}

// Optimize 执行网格搜索优化
func (g *GridOptimizer) Optimize(ctx context.Context, cfg OptimizeConfig) ([]OptimizeResult, error) {
	// 1. 生成参数组合
	combinations := g.generateCombinations(cfg.ParamSpace)
	if len(combinations) > cfg.GridMaxCombinations && cfg.GridMaxCombinations > 0 {
		combinations = combinations[:cfg.GridMaxCombinations]
	}

	if len(combinations) == 0 {
		return []OptimizeResult{}, nil
	}

	// 2. 并行执行回测
	type result struct {
		params  map[string]float64
		metrics *model.Metrics
	}

	resultsCh := make(chan result, len(combinations))
	semaphore := make(chan struct{}, 8) // 8个并发

	var wg sync.WaitGroup

	for _, params := range combinations {
		select {
		case <-ctx.Done():
			break
		default:
		}

		semaphore <- struct{}{}
		wg.Add(1)
		go func(p map[string]float64) {
			defer wg.Done()
			defer func() { <-semaphore }()

			backtestResult := runBacktest(cfg, p)
			if backtestResult != nil {
				if cfg.Constraint == nil || cfg.Constraint(*backtestResult) {
					resultsCh <- result{params: p, metrics: &backtestResult.Metrics}
				}
			}
		}(params)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]OptimizeResult, 0)
	for r := range resultsCh {
		results = append(results, OptimizeResult{
			Params:  r.params,
			Metrics: *r.metrics,
		})
	}

	// 3. 排序
	sortResults(results, cfg.Target)

	// 4. 标注rank，限制返回数量
	maxResults := 100
	if len(results) > maxResults {
		results = results[:maxResults]
	}
	for i := range results {
		results[i].Rank = i + 1
	}

	return results, nil
}

// generateCombinations 生参参数组合
func (g *GridOptimizer) generateCombinations(paramSpace map[string][]float64) []map[string]float64 {
	if len(paramSpace) == 0 {
		return nil
	}

	keys := make([]string, 0, len(paramSpace))
	values := make([][]float64, 0, len(paramSpace))
	for k, vals := range paramSpace {
		keys = append(keys, k)
		values = append(values, vals)
	}

	return g.cartesianProduct(keys, values)
}

// cartesianProduct 计算笛卡尔积
func (g *GridOptimizer) cartesianProduct(keys []string, values [][]float64) []map[string]float64 {
	if len(keys) == 0 || len(values) == 0 {
		return nil
	}

	n := len(values)
	indices := make([]int, n)
	results := make([]map[string]float64, 0)

	for {
		// 构建当前组合
		comb := make(map[string]float64)
		for i, k := range keys {
			if indices[i] < len(values[i]) {
				comb[k] = values[i][indices[i]]
			}
		}
		results = append(results, comb)

		// 递增索引
		carry := true
		for i := n - 1; i >= 0 && carry; i-- {
			indices[i]++
			if indices[i] >= len(values[i]) {
				indices[i] = 0
				carry = true
			} else {
				carry = false
			}
		}
		if carry {
			break
		}
	}

	return results
}
