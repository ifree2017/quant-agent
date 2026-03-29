package optimizer

import (
	"quant-agent/internal/model"
	"testing"
)

func TestGridOptimizer_GenerateCombinations(t *testing.T) {
	g := &GridOptimizer{}

	// 测试单参数
	ps1 := map[string][]float64{
		"MA_period": {5, 10, 20, 30},
	}
	result1 := g.generateCombinations(ps1)
	if len(result1) != 4 {
		t.Errorf("单参数期望4个组合，实际 %d", len(result1))
	}

	// 测试双参数笛卡尔积
	ps2 := map[string][]float64{
		"MA_period":  {5, 10},
		"RSI_period": {6, 14},
	}
	result2 := g.generateCombinations(ps2)
	if len(result2) != 4 {
		t.Errorf("双参数期望4个组合，实际 %d", len(result2))
	}

	// 测试三参数
	ps3 := map[string][]float64{
		"MA_period":      {5, 10},
		"RSI_period":     {6, 14},
		"RSI_overbought": {70, 80},
	}
	result3 := g.generateCombinations(ps3)
	if len(result3) != 8 {
		t.Errorf("三参数期望8个组合，实际 %d", len(result3))
	}
}

func TestGridOptimizer_Sort(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"MA_period": 10}, Metrics: model.Metrics{SharpeRatio: 0.5}},
		{Params: map[string]float64{"MA_period": 20}, Metrics: model.Metrics{SharpeRatio: 1.5}},
		{Params: map[string]float64{"MA_period": 5}, Metrics: model.Metrics{SharpeRatio: 1.0}},
	}

	sortResults(results, "sharpe_ratio")

	// 验证排序结果
	if results[0].Params["MA_period"] != 20 {
		t.Errorf("排序后第一个应该是 MA_period=20，实际 %v", results[0].Params["MA_period"])
	}
	if results[1].Params["MA_period"] != 5 {
		t.Errorf("排序后第二个应该是 MA_period=5，实际 %v", results[1].Params["MA_period"])
	}
	if results[2].Params["MA_period"] != 10 {
		t.Errorf("排序后第三个应该是 MA_period=10，实际 %v", results[2].Params["MA_period"])
	}
}

func TestGridOptimizer_SortMaxDrawdown(t *testing.T) {
	results := []OptimizeResult{
		{Params: map[string]float64{"stop_loss": 0.05}, Metrics: model.Metrics{MaxDrawdown: 0.3}},
		{Params: map[string]float64{"stop_loss": 0.02}, Metrics: model.Metrics{MaxDrawdown: 0.1}},
		{Params: map[string]float64{"stop_loss": 0.03}, Metrics: model.Metrics{MaxDrawdown: 0.2}},
	}

	sortResults(results, "max_drawdown")

	// max_drawdown 越小越好
	if results[0].Params["stop_loss"] != 0.02 {
		t.Errorf("max_drawdown排序后第一个应该是 stop_loss=0.02，实际 %v", results[0].Params["stop_loss"])
	}
	if results[2].Params["stop_loss"] != 0.05 {
		t.Errorf("max_drawdown排序后第三个应该是 stop_loss=0.05，实际 %v", results[2].Params["stop_loss"])
	}
}

func TestLatinHypercubeSample(t *testing.T) {
	space := paramSpace{
		names:  []string{"x", "y", "z"},
		bounds: [][]float64{{0, 100}, {10, 50}, {-10, 10}},
	}

	for i := 0; i < 100; i++ {
		sample := latinHypercubeSample(space)
		if len(sample) != 3 {
			t.Errorf("采样结果长度应为3，实际 %d", len(sample))
		}

		// 验证边界
		if sample[0] < 0 || sample[0] > 100 {
			t.Errorf("x应在[0,100]范围内，实际 %f", sample[0])
		}
		if sample[1] < 10 || sample[1] > 50 {
			t.Errorf("y应在[10,50]范围内，实际 %f", sample[1])
		}
		if sample[2] < -10 || sample[2] > 10 {
			t.Errorf("z应在[-10,10]范围内，实际 %f", sample[2])
		}
	}
}

func TestExtractTarget(t *testing.T) {
	m := model.Metrics{
		TotalReturn:  0.2,
		AnnualReturn: 0.15,
		SharpeRatio:  1.5,
		MaxDrawdown:  0.1,
		WinRate:      0.6,
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
		{"unknown", 1.5}, // 默认返回 sharpe_ratio
	}

	for _, tt := range tests {
		got := extractTargetValue(m, tt.target)
		if got != tt.expect {
			t.Errorf("extractTarget(%s) = %f, expect %f", tt.target, got, tt.expect)
		}
	}
}

func TestBuildParamSpace(t *testing.T) {
	ps := map[string][]float64{
		"MA_period": {5, 10, 20, 30, 60},
		"RSI_period": {6, 14, 20},
	}

	space := buildParamSpace(ps)

	if len(space.names) != 2 {
		t.Errorf("期望2个参数，实际 %d", len(space.names))
	}

	if len(space.bounds) != 2 {
		t.Errorf("期望2个边界，实际 %d", len(space.bounds))
	}

	// MA_period 边界应为 [5, 60]
	for i, name := range space.names {
		if name == "MA_period" {
			if space.bounds[i][0] != 5 || space.bounds[i][1] != 60 {
				t.Errorf("MA_period边界应为[5,60]，实际 %v", space.bounds[i])
			}
		}
	}
}

func TestOptimizeConfig_Defaults(t *testing.T) {
	cfg := OptimizeConfig{Symbol: "000001", DataDir: "./data", Days: 60}
	if cfg.Target != "" {
		t.Errorf("default target should be empty, got %s", cfg.Target)
	}
	if cfg.Method != "" {
		t.Errorf("default method should be empty, got %s", cfg.Method)
	}
	if cfg.GridMaxCombinations != 0 {
		t.Errorf("default GridMaxCombinations should be 0, got %d", cfg.GridMaxCombinations)
	}
}

func TestOptimizeResult_Rank(t *testing.T) {
	r := OptimizeResult{
		Params: map[string]float64{"MA_period": 20},
		Rank:   1,
	}
	if r.Rank != 1 {
		t.Errorf("Rank: got %d, want 1", r.Rank)
	}
	if r.Params["MA_period"] != 20 {
		t.Errorf("Params[MA_period]: got %f, want 20", r.Params["MA_period"])
	}
}

func TestBuildParamSpace_Empty(t *testing.T) {
	space := buildParamSpace(map[string][]float64{})
	if len(space.names) != 0 {
		t.Errorf("empty param space should have 0 names, got %d", len(space.names))
	}
}

func TestBuildParamSpace_SingleValue(t *testing.T) {
	ps := map[string][]float64{
		"period": {10},
	}
	space := buildParamSpace(ps)
	if len(space.names) != 1 {
		t.Errorf("single param should have 1 name, got %d", len(space.names))
	}
	if space.bounds[0][0] != 10 || space.bounds[0][1] != 10 {
		t.Errorf("single value should set min==max, got [%v, %v]", space.bounds[0][0], space.bounds[0][1])
	}
}

func TestBuildParamSpace_NoValues(t *testing.T) {
	ps := map[string][]float64{
		"period": {},
	}
	space := buildParamSpace(ps)
	if len(space.names) != 1 {
		t.Errorf("param with empty values should still have 1 name, got %d", len(space.names))
	}
	// 默认边界 [0, 100]
	if space.bounds[0][0] != 0 || space.bounds[0][1] != 100 {
		t.Errorf("empty values should default to [0,100], got [%v, %v]", space.bounds[0][0], space.bounds[0][1])
	}
}
