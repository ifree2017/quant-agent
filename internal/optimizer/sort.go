package optimizer

import (
	"quant-agent/internal/model"
	"sort"
)

// sortResults 按目标指标排序
// max_drawdown 越小越好，其他指标越大越好
func sortResults(results []OptimizeResult, target string) {
	if len(results) <= 1 {
		return
	}

	descending := true // 大多数指标越大越好
	if target == "max_drawdown" {
		descending = false // 回撤越小越好
	}

	sort.Slice(results, func(i, j int) bool {
		vi := extractTargetValue(results[i].Metrics, target)
		vj := extractTargetValue(results[j].Metrics, target)
		if descending {
			return vi > vj
		}
		return vi < vj
	})
}

// extractTargetValue 从Metrics提取目标指标值
func extractTargetValue(m model.Metrics, target string) float64 {
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
	case "calmar_ratio":
		return m.CalmarRatio
	case "sortino_ratio":
		return m.SortinoRatio
	default:
		return m.SharpeRatio
	}
}
