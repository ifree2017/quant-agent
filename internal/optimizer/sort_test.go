package optimizer

import (
	"quant-agent/internal/model"
	"testing"
)

func TestSortBySharpe(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"MA": 10}, Metrics: model.Metrics{SharpeRatio: 0.5}, Rank: 0},
		{Params: map[string]float64{"MA": 20}, Metrics: model.Metrics{SharpeRatio: 1.2}, Rank: 0},
		{Params: map[string]float64{"MA": 30}, Metrics: model.Metrics{SharpeRatio: 0.8}, Rank: 0},
	}
	sortResults(results, "sharpe_ratio")
	if results[0].Metrics.SharpeRatio != 1.2 {
		t.Errorf("first should be highest sharpe, got %.2f", results[0].Metrics.SharpeRatio)
	}
	if results[2].Metrics.SharpeRatio != 0.5 {
		t.Errorf("last should be lowest sharpe, got %.2f", results[2].Metrics.SharpeRatio)
	}
}

func TestSortByMaxDrawdown(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"stop": 0.05}, Metrics: model.Metrics{MaxDrawdown: 0.3}, Rank: 0},
		{Params: map[string]float64{"stop": 0.02}, Metrics: model.Metrics{MaxDrawdown: 0.1}, Rank: 0},
		{Params: map[string]float64{"stop": 0.03}, Metrics: model.Metrics{MaxDrawdown: 0.2}, Rank: 0},
	}
	sortResults(results, "max_drawdown")
	// max_drawdown越小越好，所以0.1应该在前面
	if results[0].Metrics.MaxDrawdown != 0.1 {
		t.Errorf("first should be smallest max_drawdown, got %.2f", results[0].Metrics.MaxDrawdown)
	}
	if results[2].Metrics.MaxDrawdown != 0.3 {
		t.Errorf("last should be largest max_drawdown, got %.2f", results[2].Metrics.MaxDrawdown)
	}
}

func TestSortByTotalReturn(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"p": 1}, Metrics: model.Metrics{TotalReturn: 0.1}, Rank: 0},
		{Params: map[string]float64{"p": 2}, Metrics: model.Metrics{TotalReturn: 0.5}, Rank: 0},
		{Params: map[string]float64{"p": 3}, Metrics: model.Metrics{TotalReturn: 0.3}, Rank: 0},
	}
	sortResults(results, "total_return")
	if results[0].Metrics.TotalReturn != 0.5 {
		t.Errorf("first should be highest total_return, got %.2f", results[0].Metrics.TotalReturn)
	}
}

func TestSortByCalmarRatio(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"p": 1}, Metrics: model.Metrics{CalmarRatio: 0.5}, Rank: 0},
		{Params: map[string]float64{"p": 2}, Metrics: model.Metrics{CalmarRatio: 1.5}, Rank: 0},
		{Params: map[string]float64{"p": 3}, Metrics: model.Metrics{CalmarRatio: 1.0}, Rank: 0},
	}
	sortResults(results, "calmar_ratio")
	if results[0].Metrics.CalmarRatio != 1.5 {
		t.Errorf("first should be highest calmar_ratio, got %.2f", results[0].Metrics.CalmarRatio)
	}
}

func TestSortBySortinoRatio(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"p": 1}, Metrics: model.Metrics{SortinoRatio: 0.8}, Rank: 0},
		{Params: map[string]float64{"p": 2}, Metrics: model.Metrics{SortinoRatio: 1.2}, Rank: 0},
		{Params: map[string]float64{"p": 3}, Metrics: model.Metrics{SortinoRatio: 1.0}, Rank: 0},
	}
	sortResults(results, "sortino_ratio")
	if results[0].Metrics.SortinoRatio != 1.2 {
		t.Errorf("first should be highest sortino_ratio, got %.2f", results[0].Metrics.SortinoRatio)
	}
}

func TestSort_Empty(t *testing.T) {
	results := []OptimizeResult{}
	sortResults(results, "sharpe_ratio")
	if len(results) != 0 {
		t.Errorf("empty slice should remain empty")
	}
}

func TestSort_SingleElement(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"p": 1}, Metrics: model.Metrics{SharpeRatio: 0.5}, Rank: 0},
	}
	sortResults(results, "sharpe_ratio")
	if results[0].Metrics.SharpeRatio != 0.5 {
		t.Errorf("single element should remain unchanged")
	}
}

func TestSort_AlreadySorted(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"p": 1}, Metrics: model.Metrics{SharpeRatio: 1.0}, Rank: 0},
		{Params: map[string]float64{"p": 2}, Metrics: model.Metrics{SharpeRatio: 0.5}, Rank: 0},
	}
	sortResults(results, "sharpe_ratio")
	if results[0].Metrics.SharpeRatio != 1.0 {
		t.Errorf("first should be 1.0, got %.2f", results[0].Metrics.SharpeRatio)
	}
}

func TestExtractTargetValue(t *testing.T) {
	m := model.Metrics{
		TotalReturn:   0.2,
		AnnualReturn:  0.15,
		SharpeRatio:   1.5,
		MaxDrawdown:   0.1,
		WinRate:       0.6,
		CalmarRatio:   0.8,
		SortinoRatio:  1.0,
	}

	tests := []struct {
		target string
		expect float64
	}{
		{"sharpe_ratio", 1.5},
		{"total_return", 0.2},
		{"max_drawdown", 0.1},
		{"annual_return", 0.15},
		{"win_rate", 0.6},
		{"calmar_ratio", 0.8},
		{"sortino_ratio", 1.0},
		{"unknown", 1.5}, // 默认返回 sharpe_ratio
	}

	for _, tt := range tests {
		got := extractTargetValue(m, tt.target)
		if got != tt.expect {
			t.Errorf("extractTargetValue(%s) = %f, expect %f", tt.target, got, tt.expect)
		}
	}
}
