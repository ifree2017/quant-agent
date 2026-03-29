package backtest

import (
	"os"
	"path/filepath"
	"quant-agent/internal/data"
	"quant-agent/internal/model"
	"testing"
	"time"
)

func TestAccount_LockUnlock(t *testing.T) {
	acc := NewAccount(100000)
	if acc.Cash != 100000 {
		t.Fatalf("initial cash: got %v, want 100000", acc.Cash)
	}

	// 锁定 20000
	acc.Lock(20000)
	if acc.Cash != 80000 {
		t.Errorf("after lock: cash got %v, want 80000", acc.Cash)
	}
	if acc.Locked != 20000 {
		t.Errorf("after lock: locked got %v, want 20000", acc.Locked)
	}

	// 再锁定 10000
	acc.Lock(10000)
	if acc.Cash != 70000 {
		t.Errorf("after 2nd lock: cash got %v, want 70000", acc.Cash)
	}
	if acc.Locked != 30000 {
		t.Errorf("after 2nd lock: locked got %v, want 30000", acc.Locked)
	}

	// 解锁 15000
	acc.Unlock(15000)
	if acc.Cash != 85000 {
		t.Errorf("after unlock: cash got %v, want 85000", acc.Cash)
	}
	if acc.Locked != 15000 {
		t.Errorf("after unlock: locked got %v, want 15000", acc.Locked)
	}

	// 再解锁 15000
	acc.Unlock(15000)
	if acc.Cash != 100000 {
		t.Errorf("after full unlock: cash got %v, want 100000", acc.Cash)
	}
	if acc.Locked != 0 {
		t.Errorf("after full unlock: locked got %v, want 0", acc.Locked)
	}
}

func TestAccount_Add(t *testing.T) {
	acc := NewAccount(100000)

	// 盈利
	acc.Add(5000)
	if acc.Cash != 105000 {
		t.Errorf("after add: got %v, want 105000", acc.Cash)
	}

	// 亏损
	acc.Add(-3000)
	if acc.Cash != 102000 {
		t.Errorf("after loss: got %v, want 102000", acc.Cash)
	}

	// 亏损到负数不应该发生（会截断到0）
	acc.Cash = 100
	acc.Add(-200)
	if acc.Cash != 0 {
		t.Errorf("after big loss: got %v, want 0", acc.Cash)
	}
}

func TestAccount_UpdatePeak(t *testing.T) {
	acc := NewAccount(100000)
	acc.UpdatePeak(100000) // 初值为初始资金
	if acc.PeakEquity != 100000 {
		t.Errorf("initial peak: got %v, want 100000", acc.PeakEquity)
	}

	acc.UpdatePeak(110000)
	if acc.PeakEquity != 110000 {
		t.Errorf("after new high: got %v, want 110000", acc.PeakEquity)
	}

	acc.UpdatePeak(105000) // 低于峰值，不更新
	if acc.PeakEquity != 110000 {
		t.Errorf("after lower equity: got %v, want 110000", acc.PeakEquity)
	}
}

func TestPortfolio_OpenClose(t *testing.T) {
	p := NewPortfolio()

	// 无持仓
	if p.HasPosition() {
		t.Error("should have no position initially")
	}
	if p.MarketValue(10.0) != 0 {
		t.Errorf("market value without position: got %v, want 0", p.MarketValue(10.0))
	}
	if p.Cost() != 0 {
		t.Errorf("cost without position: got %v, want 0", p.Cost())
	}

	// 开仓
	p.Open("2025-01-02", "000001", 10.0, 100)
	if !p.HasPosition() {
		t.Error("should have position after open")
	}
	if p.Symbol != "000001" {
		t.Errorf("symbol: got %s, want 000001", p.Symbol)
	}
	if p.EntryPriceValue() != 10.0 {
		t.Errorf("entry price: got %v, want 10.0", p.EntryPriceValue())
	}
	if p.Cost() != 1000.0 {
		t.Errorf("cost: got %v, want 1000", p.Cost())
	}

	// 市值计算
	if mv := p.MarketValue(11.0); mv != 1100.0 {
		t.Errorf("market value at 11.0: got %v, want 1100", mv)
	}

	// 平仓（盈利）
	trade := p.Close("2025-01-10", 11.0, "signal")
	if trade.PnL != 100.0 {
		t.Errorf("pnl: got %v, want 100", trade.PnL)
	}
	if trade.PnLPct != 0.1 {
		t.Errorf("pnl pct: got %v, want 0.1", trade.PnLPct)
	}
	if trade.EntryPrice != 10.0 || trade.ExitPrice != 11.0 {
		t.Errorf("prices: entry=%v, exit=%v, want 10.0, 11.0", trade.EntryPrice, trade.ExitPrice)
	}
	if p.HasPosition() {
		t.Error("should have no position after close")
	}

	// 再开仓平仓（亏损）
	p.Open("2025-01-11", "000001", 10.0, 100)
	trade2 := p.Close("2025-01-15", 9.0, "stop_loss")
	if trade2.PnL != -100.0 {
		t.Errorf("loss pnl: got %v, want -100", trade2.PnL)
	}
	if trade2.ExitReason != "stop_loss" {
		t.Errorf("exit reason: got %s, want stop_loss", trade2.ExitReason)
	}

	// 无持仓平仓返回零值
	trade3 := p.Close("2025-01-16", 9.0, "signal")
	if trade3.PnL != 0 {
		t.Errorf("zero trade pnl: got %v, want 0", trade3.PnL)
	}
}

func testDataDirBacktest(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// internal/backtest -> projectRoot
	projectRoot := filepath.Join(wd, "..", "..")
	return filepath.Join(projectRoot, "data")
}

func TestEngine_Run(t *testing.T) {
	// 使用真实 CSV 数据
	cfg := model.BacktestConfig{
		StrategyID:   "test-strategy",
		Symbol:       "000001",
		InitialCash:  100000,
		Days:         60,
	}

	dataDir := testDataDirBacktest(t)
	engine := NewEngine(cfg, dataDir)

	// 使用默认规则
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

	loader := data.NewLoader(dataDir)
	bars, err := loader.LoadBars("000001", 60)
	if err != nil {
		t.Fatalf("load bars: %v", err)
	}
	if len(bars) < 2 {
		t.Fatalf("bars too few: %d", len(bars))
	}

	result, err := engine.Run(rules, bars)
	if err != nil {
		t.Fatalf("engine run: %v", err)
	}

	if result.ID == "" {
		t.Error("result ID should not be empty")
	}
	if result.StrategyID != cfg.StrategyID {
		t.Errorf("strategy id: got %s, want %s", result.StrategyID, cfg.StrategyID)
	}
	if result.Symbol != cfg.Symbol {
		t.Errorf("symbol: got %s, want %s", result.Symbol, cfg.Symbol)
	}
	if result.InitialCash != cfg.InitialCash {
		t.Errorf("initial cash: got %v, want %v", result.InitialCash, cfg.InitialCash)
	}
	if result.Days != len(bars) {
		t.Errorf("days: got %d, want %d", result.Days, len(bars))
	}
	if len(result.EquityCurve) == 0 {
		t.Error("equity curve should not be empty")
	}
	// EquityCurve first point should have equity == initial cash (no position yet)
	if result.EquityCurve[0].Equity != cfg.InitialCash {
		t.Errorf("first equity: got %v, want %v", result.EquityCurve[0].Equity, cfg.InitialCash)
	}
}

func TestEngine_Run_BarsTooFew(t *testing.T) {
	cfg := model.BacktestConfig{
		StrategyID:   "test",
		Symbol:       "000001",
		InitialCash:  100000,
	}
	engine := NewEngine(cfg, "data")

	rules := model.StrategyRules{}
	_, err := engine.Run(rules, []data.Bar{{
		Date:  time.Now(),
		Close: 10.0,
	}})
	if err == nil {
		t.Error("expected error for too few bars")
	}
}

func TestEngine_CalcPosition(t *testing.T) {
	cfg := model.BacktestConfig{InitialCash: 100000}
	engine := NewEngine(cfg, "")

	tests := []struct {
		name   string
		rules  model.StrategyRules
		cash   float64
		price  float64
		wantMin int
		wantMax int
	}{
		{
			name:   "fixed 20%",
			rules:  model.StrategyRules{Position: model.PositionRule{Type: "fixed", Value: 0.2}},
			cash:   100000,
			price:  10.0,
			wantMin: 1900,
			wantMax: 2100,
		},
		{
			name:   "fixed 50%",
			rules:  model.StrategyRules{Position: model.PositionRule{Type: "fixed", Value: 0.5}},
			cash:   100000,
			price:  10.0,
			wantMin: 4900,
			wantMax: 5100,
		},
		{
			name:   "default 20%",
			rules:  model.StrategyRules{Position: model.PositionRule{Type: "unknown"}},
			cash:   100000,
			price:  10.0,
			wantMin: 1900,
			wantMax: 2100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.calcPosition(tt.rules, tt.cash, tt.price)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calcPosition: got %d, want between %d and %d", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
