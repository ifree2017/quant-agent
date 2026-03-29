package optimizer

import (
	"context"
	"testing"
)

func TestGridOptimizer_AllCombinations(t *testing.T) {
	g := &GridOptimizer{}
	space := map[string][]float64{
		"MA_period":  {5.0, 10.0},
		"RSI_period": {6.0, 14.0},
	}
	combinations := g.generateCombinations(space)
	// 2 x 2 = 4 combinations
	if len(combinations) != 4 {
		t.Errorf("combinations: got %d, want 4", len(combinations))
	}
}

func TestGridOptimizer_MaxCombinations(t *testing.T) {
	g := &GridOptimizer{}
	space := map[string][]float64{
		"MA_period":  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"RSI_period": {6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	}
	combinations := g.generateCombinations(space)
	if len(combinations) != 100 {
		t.Errorf("combinations: got %d, want 100", len(combinations))
	}
}

func TestGridOptimizer_EmptySpace(t *testing.T) {
	g := &GridOptimizer{}
	combinations := g.generateCombinations(map[string][]float64{})
	if combinations != nil {
		t.Errorf("empty space should return nil, got %v", combinations)
	}
}

func TestGridOptimizer_SingleParam(t *testing.T) {
	g := &GridOptimizer{}
	space := map[string][]float64{
		"period": {5.0, 10.0, 20.0},
	}
	combinations := g.generateCombinations(space)
	if len(combinations) != 3 {
		t.Errorf("single param: got %d, want 3", len(combinations))
	}
}

func TestCartesianProduct_EmptyKeys(t *testing.T) {
	g := &GridOptimizer{}
	result := g.cartesianProduct([]string{}, [][]float64{})
	if result != nil {
		t.Errorf("empty keys should return nil, got %v", result)
	}
}

func TestCartesianProduct_EmptyValues(t *testing.T) {
	g := &GridOptimizer{}
	result := g.cartesianProduct([]string{"a", "b"}, [][]float64{})
	if result != nil {
		t.Errorf("empty values should return nil, got %v", result)
	}
}

func TestCartesianProduct_ThreeParams(t *testing.T) {
	g := &GridOptimizer{}
	keys := []string{"a", "b", "c"}
	values := [][]float64{{1, 2}, {10, 20}, {100, 200}}
	result := g.cartesianProduct(keys, values)
	// 2 x 2 x 2 = 8 combinations
	if len(result) != 8 {
		t.Errorf("three params: got %d, want 8", len(result))
	}
	// Verify first combination
	if result[0]["a"] != 1 || result[0]["b"] != 10 || result[0]["c"] != 100 {
		t.Errorf("first combination wrong: %v", result[0])
	}
	// Verify last combination
	if result[7]["a"] != 2 || result[7]["b"] != 20 || result[7]["c"] != 200 {
		t.Errorf("last combination wrong: %v", result[7])
	}
}

func TestCartesianProduct_VerifyAllCombinations(t *testing.T) {
	g := &GridOptimizer{}
	space := map[string][]float64{
		"MA_period":  {5.0, 10.0},
		"RSI_period": {6.0, 14.0},
	}
	combinations := g.generateCombinations(space)

	// Should have exactly 4 combinations
	expected := map[string]float64{
		"MA_period":  5.0,
		"RSI_period": 6.0,
	}

	found := false
	for _, c := range combinations {
		if c["MA_period"] == expected["MA_period"] && c["RSI_period"] == expected["RSI_period"] {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected combination not found")
	}
}

func TestGridOptimizer_Optimize_EmptyParamSpace(t *testing.T) {
	g := &GridOptimizer{}
	cfg := OptimizeConfig{
		Symbol:      "000001",
		DataDir:     "./data",
		Days:        60,
		ParamSpace:  map[string][]float64{},
		Target:      "sharpe_ratio",
		GridMaxCombinations: 100,
	}
	
	ctx := context.Background()
	results, err := g.Optimize(ctx, cfg)
	if err != nil {
		t.Errorf("Optimize should not error on empty param space: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("empty param space should return empty results, got %d", len(results))
	}
}

func TestGridOptimizer_Optimize_GridMaxCombinationsLimit(t *testing.T) {
	g := &GridOptimizer{}
	cfg := OptimizeConfig{
		Symbol:      "000001",
		DataDir:     "./data",
		Days:        60,
		ParamSpace: map[string][]float64{
			"MA_period": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		Target:             "sharpe_ratio",
		GridMaxCombinations: 5, // Limit to 5 combinations
	}
	
	ctx := context.Background()
	results, err := g.Optimize(ctx, cfg)
	if err != nil {
		t.Errorf("Optimize should not error: %v", err)
	}
	// Even if no backtest succeeds, the function should complete
	// The limit should be applied during generation
	if len(results) > 5 {
		t.Errorf("results should be limited by GridMaxCombinations")
	}
}
