package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStyleProfile_JSON(t *testing.T) {
	now := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	p := StyleProfile{
		ID:                   "profile-1",
		UserID:               "user-1",
		Style:                "稳健",
		RiskScore:            45.5,
		TradeFrequency:       "中频",
		AvgHoldDays:          10.0,
		MaxDrawdownTolerance: 15.0,
		CreatedAt:            now,
	}

	// Marshal
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Unmarshal
	var p2 StyleProfile
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if p2.ID != p.ID {
		t.Errorf("id: got %s, want %s", p2.ID, p.ID)
	}
	if p2.Style != p.Style {
		t.Errorf("style: got %s, want %s", p2.Style, p.Style)
	}
	if p2.RiskScore != p.RiskScore {
		t.Errorf("riskScore: got %v, want %v", p2.RiskScore, p.RiskScore)
	}
	if p2.TradeFrequency != p.TradeFrequency {
		t.Errorf("tradeFrequency: got %s, want %s", p2.TradeFrequency, p.TradeFrequency)
	}
}

func TestStyleProfile_JSON_NoCreatedAt(t *testing.T) {
	p := StyleProfile{
		ID:     "p1",
		UserID: "u1",
		Style:  "激进",
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var p2 StyleProfile
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p2.Style != "激进" {
		t.Errorf("style: got %s, want 激进", p2.Style)
	}
}

func TestStrategyRules_JSON(t *testing.T) {
	rules := StrategyRules{
		Indicators: []string{"MA", "RSI", "MACD"},
		Params: map[string]float64{
			"MA_period":      20,
			"RSI_period":     14,
			"RSI_overbought": 70,
			"RSI_oversold":   30,
			"MACD_fast":       12,
			"MACD_slow":       26,
			"MACD_signal":     9,
		},
		Entry: EntryRule{
			Type:      "cross",
			Condition: "MA_cross_RSI",
		},
		Exit: ExitRule{
			Type:  "stop_loss",
			Value: 0.05,
		},
		Position: PositionRule{
			Type:  "fixed",
			Value: 0.2,
		},
		Market: "A-share",
	}

	data, err := json.Marshal(rules)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var rules2 StrategyRules
	if err := json.Unmarshal(data, &rules2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(rules2.Indicators) != 3 {
		t.Errorf("indicators count: got %d, want 3", len(rules2.Indicators))
	}
	if rules2.Entry.Type != "cross" {
		t.Errorf("entry type: got %s, want cross", rules2.Entry.Type)
	}
	if rules2.Exit.Value != 0.05 {
		t.Errorf("exit value: got %v, want 0.05", rules2.Exit.Value)
	}
	if rules2.Position.Value != 0.2 {
		t.Errorf("position value: got %v, want 0.2", rules2.Position.Value)
	}
	if rules2.Market != "A-share" {
		t.Errorf("market: got %s, want A-share", rules2.Market)
	}
	if rules2.Params["MA_period"] != 20 {
		t.Errorf("MA_period: got %v, want 20", rules2.Params["MA_period"])
	}
}

func TestStrategyRules_JSON_Empty(t *testing.T) {
	rules := StrategyRules{}

	data, err := json.Marshal(rules)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var rules2 StrategyRules
	if err := json.Unmarshal(data, &rules2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(rules2.Indicators) != 0 {
		t.Errorf("indicators: got %v, want empty", rules2.Indicators)
	}
}

func TestBacktestResult_JSON(t *testing.T) {
	now := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	result := BacktestResult{
		ID:           "bt-1",
		StrategyID:   "strategy-1",
		Symbol:       "000001",
		Days:         60,
		InitialCash:  100000,
		FinalCash:    115000,
		Metrics: Metrics{
			TotalReturn:      0.15,
			AnnualReturn:    0.25,
			SharpeRatio:     1.5,
			MaxDrawdown:     0.08,
			WinRate:         0.55,
			ProfitLossRatio: 1.8,
			TotalTrades:     20,
			CalmarRatio:     3.1,
			SortinoRatio:    2.0,
		},
		Trades: []Trade{
			{
				ID:         1,
				EntryDate:  "2025-01-02",
				ExitDate:   "2025-01-10",
				Symbol:     "000001",
				EntryPrice: 10.0,
				ExitPrice:  10.5,
				Quantity:   100,
				PnL:        50.0,
				PnLPct:     0.05,
				ExitReason: "signal",
			},
		},
		EquityCurve: []EquityPoint{
			{Date: "2025-01-02", Equity: 100000, Drawdown: 0},
			{Date: "2025-01-03", Equity: 100500, Drawdown: 0},
		},
		CreatedAt: now,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var result2 BacktestResult
	if err := json.Unmarshal(data, &result2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result2.ID != "bt-1" {
		t.Errorf("id: got %s, want bt-1", result2.ID)
	}
	if result2.FinalCash != 115000 {
		t.Errorf("finalCash: got %v, want 115000", result2.FinalCash)
	}
	if len(result2.Trades) != 1 {
		t.Errorf("trades count: got %d, want 1", len(result2.Trades))
	}
	if result2.Trades[0].PnL != 50.0 {
		t.Errorf("trade pnl: got %v, want 50.0", result2.Trades[0].PnL)
	}
	if len(result2.EquityCurve) != 2 {
		t.Errorf("equity curve count: got %d, want 2", len(result2.EquityCurve))
	}
	if result2.Metrics.TotalReturn != 0.15 {
		t.Errorf("total return: got %v, want 0.15", result2.Metrics.TotalReturn)
	}
}

func TestBacktestResult_JSON_Empty(t *testing.T) {
	result := BacktestResult{}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var result2 BacktestResult
	if err := json.Unmarshal(data, &result2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result2.Trades) != 0 {
		t.Errorf("trades: got %v, want empty", result2.Trades)
	}
}

func TestTrade_JSON(t *testing.T) {
	trade := Trade{
		ID:         1,
		EntryDate:  "2025-01-02",
		ExitDate:   "2025-01-10",
		Symbol:     "000001",
		EntryPrice: 10.0,
		ExitPrice:  11.0,
		Quantity:   100,
		PnL:        100.0,
		PnLPct:     0.10,
		ExitReason: "stop_loss",
	}

	data, err := json.Marshal(trade)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var trade2 Trade
	if err := json.Unmarshal(data, &trade2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if trade2.PnL != 100.0 {
		t.Errorf("pnl: got %v, want 100.0", trade2.PnL)
	}
	if trade2.ExitReason != "stop_loss" {
		t.Errorf("exit reason: got %s, want stop_loss", trade2.ExitReason)
	}
}

func TestMetrics_JSON(t *testing.T) {
	m := Metrics{
		TotalReturn:      0.15,
		AnnualReturn:    0.25,
		SharpeRatio:     1.5,
		MaxDrawdown:     0.08,
		WinRate:         0.55,
		ProfitLossRatio: 1.8,
		TotalTrades:     20,
		CalmarRatio:     3.1,
		SortinoRatio:    2.0,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m2 Metrics
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m2.SharpeRatio != 1.5 {
		t.Errorf("sharpe ratio: got %v, want 1.5", m2.SharpeRatio)
	}
	if m2.WinRate != 0.55 {
		t.Errorf("win rate: got %v, want 0.55", m2.WinRate)
	}
}

func TestEquityPoint_JSON(t *testing.T) {
	ep := EquityPoint{
		Date:     "2025-01-02",
		Equity:   105000,
		Drawdown: 0.05,
	}

	data, err := json.Marshal(ep)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var ep2 EquityPoint
	if err := json.Unmarshal(data, &ep2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ep2.Equity != 105000 {
		t.Errorf("equity: got %v, want 105000", ep2.Equity)
	}
	if ep2.Drawdown != 0.05 {
		t.Errorf("drawdown: got %v, want 0.05", ep2.Drawdown)
	}
}

func TestStrategy_JSON(t *testing.T) {
	now := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	s := Strategy{
		ID:        "s1",
		UserID:    "u1",
		Name:      "测试策略",
		Style:     "稳健",
		Rules: StrategyRules{
			Indicators: []string{"MA"},
			Params:     map[string]float64{"MA_period": 5},
		},
		Version:   1,
		Tags:      []string{"tag1", "tag2"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var s2 Strategy
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if s2.Name != "测试策略" {
		t.Errorf("name: got %s, want 测试策略", s2.Name)
	}
	if len(s2.Tags) != 2 {
		t.Errorf("tags count: got %d, want 2", len(s2.Tags))
	}
}

func TestBacktestConfig_JSON(t *testing.T) {
	cfg := BacktestConfig{
		StrategyID:   "s1",
		Symbol:      "000001",
		Days:        60,
		InitialCash: 50000,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var cfg2 BacktestConfig
	if err := json.Unmarshal(data, &cfg2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if cfg2.InitialCash != 50000 {
		t.Errorf("initialCash: got %v, want 50000", cfg2.InitialCash)
	}
}

func TestTradeRecord_JSON(t *testing.T) {
	tr := TradeRecord{
		Date:     "2025-01-02",
		Action:   "buy",
		Symbol:   "000001",
		Price:    10.5,
		Quantity: 100,
	}

	data, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var tr2 TradeRecord
	if err := json.Unmarshal(data, &tr2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if tr2.Action != "buy" {
		t.Errorf("action: got %s, want buy", tr2.Action)
	}
	if tr2.Price != 10.5 {
		t.Errorf("price: got %v, want 10.5", tr2.Price)
	}
}

func TestStyleAnalyzeRequest_JSON(t *testing.T) {
	req := StyleAnalyzeRequest{
		UserID: "u1",
		Records: []TradeRecord{
			{Date: "2025-01-02", Action: "buy", Symbol: "000001", Price: 10.0, Quantity: 100},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var req2 StyleAnalyzeRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req2.UserID != "u1" {
		t.Errorf("userID: got %s, want u1", req2.UserID)
	}
	if len(req2.Records) != 1 {
		t.Errorf("records count: got %d, want 1", len(req2.Records))
	}
}

func TestStrategyGenerateRequest_JSON(t *testing.T) {
	req := StrategyGenerateRequest{
		UserID: "u1",
		StyleProfile: StyleProfile{
			Style:     "激进",
			RiskScore: 70,
		},
		Market: "A-share",
		Symbol: "000001",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var req2 StrategyGenerateRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req2.StyleProfile.Style != "激进" {
		t.Errorf("style: got %s, want 激进", req2.StyleProfile.Style)
	}
	if req2.Market != "A-share" {
		t.Errorf("market: got %s, want A-share", req2.Market)
	}
}

func TestBacktestRunRequest_JSON(t *testing.T) {
	req := BacktestRunRequest{
		StrategyID:  "s1",
		Symbol:     "000001",
		Days:       60,
		InitialCash: 100000,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var req2 BacktestRunRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req2.StrategyID != "s1" {
		t.Errorf("strategyID: got %s, want s1", req2.StrategyID)
	}
	if req2.Days != 60 {
		t.Errorf("days: got %d, want 60", req2.Days)
	}
}
