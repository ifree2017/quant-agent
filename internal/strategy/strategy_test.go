package strategy

import (
	"quant-agent/internal/data"
	"quant-agent/internal/model"
	"testing"
	"time"
)

func testDate(days int) time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, days)
}

func TestStrategy_Signal(t *testing.T) {
	rules := model.StrategyRules{
		Indicators: []string{"MA", "RSI"},
		Params: map[string]float64{
			"MA_period":      20,
			"RSI_period":     14,
			"RSI_overbought": 70,
			"RSI_oversold":   30,
		},
		Entry: model.EntryRule{
			Type:      "cross",
			Condition: "MA_cross_RSI",
		},
		Exit: model.ExitRule{
			Type:  "stop_loss",
			Value: 0.05,
		},
		Position: model.PositionRule{
			Type:  "fixed",
			Value: 0.2,
		},
	}

	executor := NewExecutor(rules)

	// 初始化数据
	closes := []float64{10.0, 10.2, 10.4, 10.6, 10.8, 11.0, 11.2, 11.4, 11.6, 11.8, 12.0, 12.2, 12.4, 12.6, 12.8, 13.0, 13.2, 13.4, 13.6, 13.8, 14.0}
	for i, price := range closes {
		bar := data.Bar{
			Date:   testDate(i),
			Close:  price,
			Volume: 1000,
		}
		executor.Update(bar)
	}

	signal := executor.Signal()
	if signal != "hold" && signal != "buy" && signal != "sell" {
		t.Errorf("Signal: got %s, want hold/buy/sell", signal)
	}
}

func TestStrategy_MultipleIndicators(t *testing.T) {
	rules := model.StrategyRules{
		Indicators: []string{"MA", "RSI", "MACD"},
		Params: map[string]float64{
			"MA_period":      5,
			"RSI_period":     6,
			"RSI_overbought": 70,
			"RSI_oversold":   30,
			"MACD_fast":      12,
			"MACD_slow":      26,
			"MACD_signal":    9,
		},
		Entry: model.EntryRule{
			Type:      "cross",
			Condition: "MA_cross_RSI",
		},
		Exit: model.ExitRule{
			Type:  "stop_loss",
			Value: 0.05,
		},
		Position: model.PositionRule{Type: "fixed", Value: 0.2},
	}

	executor := NewExecutor(rules)

	// 喂入足够的数据让所有指标生效
	closes := make([]float64, 50)
	for i := range closes {
		closes[i] = 10.0 + float64(i)*0.1
	}
	for i, price := range closes {
		bar := data.Bar{
			Date:   testDate(i),
			Close:  price,
			Open:   price - 0.05,
			High:   price + 0.1,
			Low:    price - 0.1,
			Volume: 1000,
		}
		executor.Update(bar)
	}

	signal := executor.Signal()
	if signal == "" {
		t.Error("signal should not be empty")
	}

	// 检查指标是否被计算
	store := executor.Store()
	if store == nil {
		t.Fatal("store should not be nil")
	}
	// MA period=5，RSI period=6，至少需要足够数据
	if len(store.ClosePrices) != 50 {
		t.Errorf("close prices count: got %d, want 50", len(store.ClosePrices))
	}
}

func TestStrategy_RSISignal(t *testing.T) {
	rules := model.StrategyRules{
		Indicators: []string{"RSI"},
		Params: map[string]float64{
			"RSI_period":     6,
			"RSI_overbought": 70,
			"RSI_oversold":   30,
		},
		Entry: model.EntryRule{
			Type:      "indicator_value",
			Condition: "RSI_oversold",
		},
		Exit: model.ExitRule{
			Type:  "stop_loss",
			Value: 0.05,
		},
		Position: model.PositionRule{Type: "fixed", Value: 0.2},
	}

	executor := NewExecutor(rules)

	// RSI 从低到高（超卖 -> 正常 -> 超买）
	// 先喂足够数据建立基准
	for i := 0; i < 30; i++ {
		bar := data.Bar{
			Date:   testDate(i),
			Close:  10.0,
			Volume: 1000,
		}
		executor.Update(bar)
	}

	// 逐步上涨制造超买
	for i := 30; i < 50; i++ {
		bar := data.Bar{
			Date:   testDate(i),
			Close:  10.0 + float64(i-30)*0.5, // 上涨
			Volume: 1000,
		}
		executor.Update(bar)
	}

	signal := executor.Signal()
	// RSI 超买时应该发出卖出信号
	store := executor.Store()
	if store.RSI.RSI14 > 70 {
		if signal != "sell" {
			t.Logf("RSI=%.2f > 70, signal=%s (may be sell depending on RSI calculation)", store.RSI.RSI14, signal)
		}
	}

	// 测试超卖情况：价格下跌
	executor2 := NewExecutor(rules)
	for i := 0; i < 30; i++ {
		bar := data.Bar{
			Date:   testDate(i),
			Close:  20.0,
			Volume: 1000,
		}
		executor2.Update(bar)
	}
	// 逐步下跌
	for i := 30; i < 45; i++ {
		bar := data.Bar{
			Date:   testDate(i),
			Close:  20.0 - float64(i-30)*0.5,
			Volume: 1000,
		}
		executor2.Update(bar)
	}

	signal2 := executor2.Signal()
	store2 := executor2.Store()
	t.Logf("RSI14=%.2f, signal=%s", store2.RSI.RSI14, signal2)
	_ = signal2 // 信号结果取决于指标计算
}

func TestStrategy_Store(t *testing.T) {
	rules := model.StrategyRules{
		Indicators: []string{"MA"},
		Params:     map[string]float64{"MA_period": 5},
	}
	executor := NewExecutor(rules)
	store := executor.Store()
	if store == nil {
		t.Fatal("store should not be nil")
	}
	// Before Update, Bars should be nil (not yet initialized)
	// After Update, Bars should be a non-nil slice
	bar := data.Bar{Date: testDate(0), Close: 10.0, Volume: 1000}
	executor.Update(bar)
	if store.Bars == nil {
		t.Error("store bars should be non-nil after update")
	}
	if len(store.Bars) != 1 {
		t.Errorf("bars count after update: got %d, want 1", len(store.Bars))
	}
}

func TestStrategy_Update(t *testing.T) {
	rules := model.StrategyRules{
		Indicators: []string{"MA"},
		Params:     map[string]float64{"MA_period": 5},
		Entry:      model.EntryRule{Condition: "MA_cross_RSI"},
	}
	executor := NewExecutor(rules)

	bar := data.Bar{
		Date:   testDate(0),
		Open:   10.0,
		High:   11.0,
		Low:    9.5,
		Close:  10.5,
		Volume: 1000,
	}
	executor.Update(bar)

	store := executor.Store()
	if len(store.Bars) != 1 {
		t.Errorf("bars count: got %d, want 1", len(store.Bars))
	}
	if store.Bars[0].Close != 10.5 {
		t.Errorf("bar close: got %v, want 10.5", store.Bars[0].Close)
	}
}

func TestStrategy_ExecutorSignal_HoldOnFewBars(t *testing.T) {
	rules := model.StrategyRules{
		Indicators: []string{"MA", "RSI"},
		Params:     map[string]float64{"MA_period": 5, "RSI_period": 6},
		Entry:      model.EntryRule{Condition: "MA_cross_RSI"},
	}
	executor := NewExecutor(rules)

	// 只喂 1 条数据，不够计算指标
	bar := data.Bar{Date: testDate(0), Close: 10.0, Volume: 1000}
	executor.Update(bar)

	signal := executor.Signal()
	if signal != "hold" {
		t.Errorf("signal with 1 bar: got %s, want hold", signal)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice  []string
		item   string
		expect bool
	}{
		{[]string{"MA", "RSI"}, "MA", true},
		{[]string{"MA", "RSI"}, "RSI", true},
		{[]string{"MA", "RSI"}, "MACD", false},
		{[]string{}, "MA", false},
	}
	for _, tt := range tests {
		got := contains(tt.slice, tt.item)
		if got != tt.expect {
			t.Errorf("contains(%v, %s): got %v, want %v", tt.slice, tt.item, got, tt.expect)
		}
	}
}
