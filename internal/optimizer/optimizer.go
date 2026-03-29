package optimizer

import (
	"context"
	"quant-agent/internal/backtest"
	"quant-agent/internal/data"
	"quant-agent/internal/model"
)

// Optimizer 参数优化器接口
type Optimizer interface {
	// Optimize 执行参数优化
	Optimize(ctx context.Context, cfg OptimizeConfig) ([]OptimizeResult, error)
}

// OptimizeConfig 优化配置
type OptimizeConfig struct {
	Symbol      string
	DataDir     string
	Days        int
	StrategyRules model.StrategyRules // 基础规则模板

	// 待优化参数空间
	ParamSpace map[string][]float64 // 参数名 → 候选值列表
	Target     string               // 优化目标：sharpe_ratio / total_return / max_drawdown

	// 约束
	Constraint func(result model.BacktestResult) bool // 过滤条件

	// 优化方法
	Method string // "grid" | "bayesian"

	// 网格搜索专用
	GridMaxCombinations int // 最大组合数（避免爆炸）
}

// OptimizeResult 优化结果
type OptimizeResult struct {
	Params  map[string]float64
	Metrics model.Metrics
	Rank    int
}

// paramSpace 构建参数空间（用于贝叶斯优化）
type paramSpace struct {
	names  []string
	bounds [][]float64 // [param][min, max]
}

// buildParamSpace 从ParamSpace构建paramSpace
func buildParamSpace(ps map[string][]float64) paramSpace {
	space := paramSpace{
		names:  make([]string, 0, len(ps)),
		bounds: make([][]float64, 0, len(ps)),
	}
	for k, vals := range ps {
		space.names = append(space.names, k)
		if len(vals) >= 2 {
			min, max := vals[0], vals[len(vals)-1]
			for _, v := range vals {
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
			}
			space.bounds = append(space.bounds, []float64{min, max})
		} else if len(vals) == 1 {
			space.bounds = append(space.bounds, []float64{vals[0], vals[0]})
		} else {
			space.bounds = append(space.bounds, []float64{0, 100})
		}
	}
	return space
}

// runBacktest 在参数组合下执行回测
func runBacktest(cfg OptimizeConfig, params map[string]float64) *model.BacktestResult {
	loader := data.NewLoader(cfg.DataDir)
	bars, err := loader.LoadBars(cfg.Symbol, cfg.Days)
	if err != nil || len(bars) < 2 {
		return nil
	}

	rules := cfg.StrategyRules
	for k, v := range params {
		rules.Params[k] = v
	}

	engine := backtest.NewEngine(model.BacktestConfig{
		Symbol: cfg.Symbol,
		Days:   cfg.Days,
	}, cfg.DataDir)
	result, err := engine.Run(rules, bars)
	if err != nil {
		return nil
	}
	return result
}
