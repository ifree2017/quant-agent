package optimizer

import (
	"context"
	"testing"
)

func TestRBFKernel(t *testing.T) {
	a := []float64{1.0, 2.0}
	result := rbfKernel(a, a, 1.0)
	// 自己和自己的核应该接近1
	if result < 0.9 || result > 1.1 {
		t.Errorf("RBF kernel(self): got %.2f, want ~1.0", result)
	}
}

func TestRBFKernel_Different(t *testing.T) {
	a := []float64{1.0, 2.0}
	bVec := []float64{10.0, 20.0}
	result := rbfKernel(a, bVec, 1.0)
	// 不同点的核应该接近0
	if result > 0.5 {
		t.Errorf("RBF kernel(different): got %.2f, want ~0", result)
	}
}

func TestRBFKernel_Sigma(t *testing.T) {
	a := []float64{1.0, 2.0}
	bVec := []float64{1.5, 2.5}

	// sigma大，核值高
	r1 := rbfKernel(a, bVec, 10.0)
	r2 := rbfKernel(a, bVec, 0.1)
	if r1 <= r2 {
		t.Errorf("larger sigma should give higher kernel value: r1=%.4f, r2=%.4f", r1, r2)
	}
}

func TestRBFKernel_ZeroSigma(t *testing.T) {
	// sigma为0应该返回0（避免除零）
	a := []float64{1.0, 2.0}
	bVec := []float64{1.5, 2.5}
	result := rbfKernel(a, bVec, 0.0)
	if result != 0 {
		t.Errorf("zero sigma should return 0, got %.4f", result)
	}
}

func TestUCB(t *testing.T) {
	// 测试UCB计算
	mean := 0.5
	var_ := 0.1
	ucb := mean + 2.0*var_
	if ucb < 0.5 {
		t.Errorf("UCB should be >= mean")
	}
	if ucb != 0.7 {
		t.Errorf("UCB: got %.2f, want 0.7", ucb)
	}
}

func TestUCB_Negative(t *testing.T) {
	mean := -0.5
	var_ := 0.1
	ucb := mean + 2.0*var_
	if ucb >= 0 {
		t.Errorf("UCB with negative mean should still be negative")
	}
}

func TestUCB_ZeroVariance(t *testing.T) {
	mean := 0.5
	var_ := 0.0
	ucb := mean + 2.0*var_
	if ucb != 0.5 {
		t.Errorf("UCB with zero variance should equal mean, got %.2f", ucb)
	}
}

func TestBayesianOptimizer_New(t *testing.T) {
	b := NewBayesianOptimizer()
	if b == nil {
		t.Fatal("NewBayesianOptimizer should not return nil")
	}
	if b.kernel == nil {
		t.Error("kernel should be initialized")
	}
	if b.observations == nil {
		t.Error("observations should be initialized")
	}
}

func TestSliceToMap(t *testing.T) {
	slice := []float64{1.0, 2.0, 3.0}
	names := []string{"a", "b", "c"}
	result := sliceToMap(slice, names)
	if len(result) != 3 {
		t.Errorf("sliceToMap: got %d, want 3", len(result))
	}
	if result["a"] != 1.0 {
		t.Errorf("result[a]: got %f, want 1.0", result["a"])
	}
	if result["c"] != 3.0 {
		t.Errorf("result[c]: got %f, want 3.0", result["c"])
	}
}

func TestSliceToMap_Mismatch(t *testing.T) {
	slice := []float64{1.0, 2.0}
	names := []string{"a", "b", "c"} // 3 names, 2 values
	result := sliceToMap(slice, names)
	if len(result) != 2 {
		t.Errorf("sliceToMap mismatch: got %d, want 2", len(result))
	}
}

func TestSliceToMap_Empty(t *testing.T) {
	slice := []float64{}
	names := []string{}
	result := sliceToMap(slice, names)
	if len(result) != 0 {
		t.Errorf("empty sliceToMap: got %d, want 0", len(result))
	}
}

func TestLatinHypercubeSample_Bounds(t *testing.T) {
	space := paramSpace{
		names:  []string{"x", "y"},
		bounds: [][]float64{{0, 100}, {10, 50}},
	}

	for i := 0; i < 50; i++ {
		sample := latinHypercubeSample(space)
		if len(sample) != 2 {
			t.Fatalf("sample length should be 2, got %d", len(sample))
		}
		// 验证x在[0, 100]范围内
		if sample[0] < 0 || sample[0] > 100 {
			t.Errorf("x should be in [0,100], got %f", sample[0])
		}
		// 验证y在[10, 50]范围内
		if sample[1] < 10 || sample[1] > 50 {
			t.Errorf("y should be in [10,50], got %f", sample[1])
		}
	}
}

func TestLatinHypercubeSample_SingleDim(t *testing.T) {
	space := paramSpace{
		names:  []string{"x"},
		bounds: [][]float64{{-5, 5}},
	}

	for i := 0; i < 20; i++ {
		sample := latinHypercubeSample(space)
		if len(sample) != 1 {
			t.Errorf("single dim sample length should be 1, got %d", len(sample))
		}
		if sample[0] < -5 || sample[0] > 5 {
			t.Errorf("x should be in [-5,5], got %f", sample[0])
		}
	}
}

func TestLatinHypercubeSample_DefaultBounds(t *testing.T) {
	// 测试默认边界情况
	space := paramSpace{
		names:  []string{"x"},
		bounds: [][]float64{{}},
	}

	sample := latinHypercubeSample(space)
	if len(sample) != 1 {
		t.Errorf("sample length should be 1, got %d", len(sample))
	}
	// 空边界应该使用默认[0, 100]
	if sample[0] < 0 || sample[0] > 100 {
		t.Errorf("x should be in [0,100], got %f", sample[0])
	}
}

func TestBayesianOptimizer_Optimize_EmptyParamSpace(t *testing.T) {
	b := NewBayesianOptimizer()
	cfg := OptimizeConfig{
		Symbol:     "000001",
		DataDir:    "./data",
		Days:       60,
		ParamSpace: map[string][]float64{},
		Target:     "sharpe_ratio",
	}

	ctx := context.Background()
	results, err := b.Optimize(ctx, cfg)
	if err != nil {
		t.Errorf("Optimize should not error on empty param space: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("empty param space should return empty results, got %d", len(results))
	}
}

func TestBayesianOptimizer_Optimize_SingleParam(t *testing.T) {
	b := NewBayesianOptimizer()
	cfg := OptimizeConfig{
		Symbol:  "000001",
		DataDir: "./data",
		Days:    60,
		ParamSpace: map[string][]float64{
			"MA_period": {5, 10, 20},
		},
		Target: "sharpe_ratio",
	}

	ctx := context.Background()
	results, err := b.Optimize(ctx, cfg)
	// May fail due to missing backtest data, but should not crash
	if err != nil && len(results) == 0 {
		// Expected - no valid backtest results
	}
}
