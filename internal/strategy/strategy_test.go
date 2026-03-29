package strategy

import (
	"testing"
	"time"
	"quant-agent/internal/model"
	"quant-agent/internal/data"
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
