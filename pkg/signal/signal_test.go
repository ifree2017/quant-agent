package signal

import (
	"testing"
	"time"
)

func TestSignalTypeConstants(t *testing.T) {
	if SignalBuy != "BUY" {
		t.Errorf("SignalBuy: got %s, want BUY", SignalBuy)
	}
	if SignalSell != "SELL" {
		t.Errorf("SignalSell: got %s, want SELL", SignalSell)
	}
	if SignalHold != "HOLD" {
		t.Errorf("SignalHold: got %s, want HOLD", SignalHold)
	}
	if SignalWatch != "WATCH" {
		t.Errorf("SignalWatch: got %s, want WATCH", SignalWatch)
	}
}

func TestAlertLevelConstants(t *testing.T) {
	if AlertLow != "LOW" {
		t.Errorf("AlertLow: got %s, want LOW", AlertLow)
	}
	if AlertMedium != "MEDIUM" {
		t.Errorf("AlertMedium: got %s, want MEDIUM", AlertMedium)
	}
	if AlertHigh != "HIGH" {
		t.Errorf("AlertHigh: got %s, want HIGH", AlertHigh)
	}
	if AlertCritical != "CRITICAL" {
		t.Errorf("AlertCritical: got %s, want CRITICAL", AlertCritical)
	}
}

func TestPlatformWeight(t *testing.T) {
	total := 0.0
	for platform, weight := range PlatformWeight {
		if weight <= 0 || weight > 1 {
			t.Errorf("PlatformWeight[%s]: got %f, want between 0 and 1", platform, weight)
		}
		total += weight
	}

	// 权重总和应接近1
	if total < 0.99 || total > 1.01 {
		t.Errorf("PlatformWeight total: got %f, want ~1.0", total)
	}
}

func TestScoreThresholds(t *testing.T) {
	if ScoreThresholds.BuyScore <= 0 {
		t.Errorf("BuyScore should be positive, got %f", ScoreThresholds.BuyScore)
	}
	if ScoreThresholds.SellScore >= 0 {
		t.Errorf("SellScore should be negative, got %f", ScoreThresholds.SellScore)
	}
	if ScoreThresholds.BuyConf <= 0 || ScoreThresholds.BuyConf > 1 {
		t.Errorf("BuyConf should be between 0 and 1, got %f", ScoreThresholds.BuyConf)
	}
}

func TestSectorKeywords(t *testing.T) {
	if len(SectorKeywords) == 0 {
		t.Error("SectorKeywords should not be empty")
	}

	// 检查关键板块
	expectedSectors := []string{"新能源", "半导体", "人工智能", "医药", "消费"}
	for _, sector := range expectedSectors {
		if _, ok := SectorKeywords[sector]; !ok {
			t.Errorf("Expected sector %s not found", sector)
		}
	}

	// 检查每个板块都有关键词
	for sector, keywords := range SectorKeywords {
		if len(keywords) == 0 {
			t.Errorf("Sector %s has no keywords", sector)
		}
	}
}

func TestNegativeKeywords(t *testing.T) {
	if len(NegativeKeywords) == 0 {
		t.Error("NegativeKeywords should not be empty")
	}

	// 检查关键负面词
	expected := []string{"黑天鹅", "减持", "解禁", "业绩下滑", "监管"}
	for _, kw := range expected {
		found := false
		for _, nk := range NegativeKeywords {
			if nk == kw {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected negative keyword %s not found", kw)
		}
	}
}

func TestSignal_Structure(t *testing.T) {
	s := Signal{
		StockCode:  "000001",
		SignalType: SignalBuy,
		Confidence: 0.85,
		Source:     "social_sentiment",
		Score:      72.5,
		Weight:     1.0,
		Message:    "Test message",
		CreatedAt:  time.Now(),
	}

	if s.StockCode != "000001" {
		t.Errorf("StockCode: got %s, want 000001", s.StockCode)
	}
	if s.SignalType != SignalBuy {
		t.Errorf("SignalType: got %s, want BUY", s.SignalType)
	}
	if s.Confidence != 0.85 {
		t.Errorf("Confidence: got %f, want 0.85", s.Confidence)
	}
	if s.Score != 72.5 {
		t.Errorf("Score: got %f, want 72.5", s.Score)
	}
}

func TestSectorSignal_Structure(t *testing.T) {
	ss := SectorSignal{
		SectorName: "新能源",
		HeatScore:  85.5,
		Direction:  "emerging",
		StockCodes: []string{"000001", "000002"},
		Keywords:   []string{"锂电", "光伏"},
		ChangePct:  5.2,
		Source:     "sector_rotation",
		CreatedAt:  time.Now(),
	}

	if ss.SectorName != "新能源" {
		t.Errorf("SectorName: got %s, want 新能源", ss.SectorName)
	}
	if ss.Direction != "emerging" {
		t.Errorf("Direction: got %s, want emerging", ss.Direction)
	}
	if len(ss.StockCodes) != 2 {
		t.Errorf("StockCodes length: got %d, want 2", len(ss.StockCodes))
	}
}

func TestAlertSignal_Structure(t *testing.T) {
	as := AlertSignal{
		StockCode:   "000001",
		AlertLevel:  AlertHigh,
		Title:       "High risk alert",
		Description: "Multiple negative signals detected",
		Keywords:    []string{"减持", "监管"},
		NewsCount:   5,
		Duration:    "中期",
		Source:      "news_alert",
		CreatedAt:   time.Now(),
	}

	if as.AlertLevel != AlertHigh {
		t.Errorf("AlertLevel: got %s, want HIGH", as.AlertLevel)
	}
	if as.NewsCount != 5 {
		t.Errorf("NewsCount: got %d, want 5", as.NewsCount)
	}
}

func TestMaxMin(t *testing.T) {
	if max(3, 5) != 5 {
		t.Errorf("max(3, 5): got 5, want 5")
	}
	if max(7, 2) != 7 {
		t.Errorf("max(7, 2): got 7, want 7")
	}

	if min(3, 5) != 3 {
		t.Errorf("min(3, 5): got 3, want 3")
	}
	if min(7, 2) != 2 {
		t.Errorf("min(7, 2): got 2, want 2")
	}
}

func TestUnique(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	result := unique(input)

	if len(result) != 4 {
		t.Errorf("unique length: got %d, want 4", len(result))
	}

	// 检查去重
	seen := make(map[string]bool)
	for _, s := range result {
		if seen[s] {
			t.Errorf("unique did not remove duplicate: %s", s)
		}
		seen[s] = true
	}
}
