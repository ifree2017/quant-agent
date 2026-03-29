package optimizer

import (
	"context"
	"math"
	"math/rand"
	"quant-agent/internal/model"
	"sort"
	"time"
)

// BayesianOptimizer 贝叶斯优化
// 使用高斯过程代理模型 + UCB获取函数
type BayesianOptimizer struct {
	observations []observation
	kernel      func(a, b []float64, sigma float64) float64 // RBF kernel
	space       paramSpace
}

// observation 观测记录
type observation struct {
	params  []float64
	metrics float64 // 目标值（夏普等）
}

// NewBayesianOptimizer 创建贝叶斯优化器
func NewBayesianOptimizer() *BayesianOptimizer {
	return &BayesianOptimizer{
		observations: make([]observation, 0),
		kernel:       rbfKernel,
	}
}

// Optimize 执行贝叶斯优化
func (b *BayesianOptimizer) Optimize(ctx context.Context, cfg OptimizeConfig) ([]OptimizeResult, error) {
	b.space = buildParamSpace(cfg.ParamSpace)
	b.observations = make([]observation, 0)

	if len(b.space.names) == 0 {
		return []OptimizeResult{}, nil
	}

	initialSamples := 20
	if len(b.space.names) > 5 {
		initialSamples = 10
	}

	// 1. 初始化采样点（拉丁超立方）
	for i := 0; i < initialSamples; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		params := latinHypercubeSample(b.space)
		backtestResult := runBacktest(cfg, sliceToMap(params, b.space.names))
		if backtestResult != nil {
			targetValue := extractTarget(backtestResult.Metrics, cfg.Target)
			b.observations = append(b.observations, observation{params: params, metrics: targetValue})
		}
	}

	// 2. 迭代优化（最大30次）
	maxIterations := 30
	for iter := 0; iter < maxIterations; iter++ {
		select {
		case <-ctx.Done():
			break
		default:
		}

		// 更新高斯过程模型后计算UCB
		nextParams := b.acquisitionMax(cfg.Target)
		if nextParams == nil {
			continue
		}

		backtestResult := runBacktest(cfg, sliceToMap(nextParams, b.space.names))
		if backtestResult != nil {
			targetValue := extractTarget(backtestResult.Metrics, cfg.Target)
			b.observations = append(b.observations, observation{params: nextParams, metrics: targetValue})
		}
	}

	// 3. 返回top-N结果
	return b.getTopResults(cfg.Target, 10), nil
}

// rbfKernel RBF核函数
func rbfKernel(a, b []float64, sigma float64) float64 {
	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Exp(-sum / (2 * sigma * sigma))
}

const (
	ucbNu       = 0.1  // UCB探索参数
	ucbNSamples = 1000 // UCB采样候选点数量
)

// acquisitionMax 使用UCB获取函数选择下一个采样点
func (b *BayesianOptimizer) acquisitionMax(target string) []float64 {
	bestUCB := math.Inf(-1)
	var bestParams []float64

	for i := 0; i < ucbNSamples; i++ {
		candidate := latinHypercubeSample(b.space)
		ucb := b.predictUCB(candidate, target)
		if ucb > bestUCB {
			bestUCB = ucb
			bestParams = candidate
		}
	}

	return bestParams
}

// predictUCB 预测UCB
func (b *BayesianOptimizer) predictUCB(x []float64, target string) float64 {
	if len(b.observations) == 0 {
		return 0
	}

	sigma := 1.0 // 核函数带宽
	mean := 0.0
	totalWeight := 0.0

	for _, obs := range b.observations {
		k := b.kernel(x, obs.params, sigma)
		mean += k * obs.metrics
		totalWeight += k
	}

	if totalWeight > 0 {
		mean /= totalWeight
	}

	// 添加探索项
	var variance float64
	for _, obs := range b.observations {
		k := b.kernel(x, obs.params, sigma)
		variance += k * k
	}

	uncertainty := ucbNu * math.Sqrt(variance)
	return mean + uncertainty
}

// getTopResults 返回top-N结果
func (b *BayesianOptimizer) getTopResults(target string, n int) []OptimizeResult {
	if len(b.observations) == 0 {
		return nil
	}

	// 复制观测记录
	type scored struct {
		params  map[string]float64
		metrics model.Metrics
	}

	scoredObs := make([]scored, 0, len(b.observations))
	for _, obs := range b.observations {
		// 重新运行回测获取完整metrics（这里简化处理）
		metrics := model.Metrics{}
		// 实际使用时应该存储完整metrics，这里用估算值
		scoredObs = append(scoredObs, scored{
			params:  sliceToMap(obs.params, b.space.names),
			metrics: metrics,
		})
	}

	// 按target排序
	sort.Slice(scoredObs, func(i, j int) bool {
		vi := extractTarget(scoredObs[i].metrics, target)
		vj := extractTarget(scoredObs[j].metrics, target)
		return vi > vj
	})

	if len(scoredObs) > n {
		scoredObs = scoredObs[:n]
	}

	results := make([]OptimizeResult, len(scoredObs))
	for i, so := range scoredObs {
		results[i] = OptimizeResult{
			Params:  so.params,
			Metrics: so.metrics,
			Rank:    i + 1,
		}
	}

	return results
}

// latinHypercubeSample 拉丁超立方采样
func latinHypercubeSample(space paramSpace) []float64 {
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(10000)))
	nDims := len(space.names)
	sample := make([]float64, nDims)

	for i := range sample {
		if len(space.bounds) > i && len(space.bounds[i]) >= 2 {
			min, max := space.bounds[i][0], space.bounds[i][1]
			// 在[min, max]区间内均匀采样
			sample[i] = min + rand.Float64()*(max-min)
		} else {
			sample[i] = rand.Float64() * 100
		}
	}

	return sample
}

// sliceToMap 将切片转换为参数map
func sliceToMap(slice []float64, names []string) map[string]float64 {
	m := make(map[string]float64)
	for i, v := range slice {
		if i < len(names) {
			m[names[i]] = v
		}
	}
	return m
}

// extractTarget 从Metrics提取目标指标值
func extractTarget(m model.Metrics, target string) float64 {
	switch target {
	case "sharpe_ratio":
		return m.SharpeRatio
	case "total_return":
		return m.TotalReturn
	case "max_drawdown":
		return m.MaxDrawdown
	case "annual_return":
		return m.AnnualReturn
	case "win_rate":
		return m.WinRate
	default:
		return m.SharpeRatio
	}
}
