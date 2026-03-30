package signal

import (
	"testing"
	"time"
)

func TestBuyPointPatterns_ScoreCalculation(t *testing.T) {
	tests := []struct {
		name      string
		patterns  BuyPointPatterns
		wantScore int
		wantLevel BuyPointLevel
		wantAction TradeAction
	}{
		{
			name: "STRONG_BUY - all patterns positive",
			patterns: BuyPointPatterns{
				Pullback:   true,
				Breakout:   true,
				SecondWave: true,
				Flow:       true,
				Pressure:   true,
			},
			wantScore:  11, // 2+3+3+2+1
			wantLevel:  BuyPointStrong,
			wantAction: ActionBuyAll,
		},
		{
			name: "STRONG_BUY - score >= 6",
			patterns: BuyPointPatterns{
				Pullback:   true,
				Breakout:   true,
				SecondWave: false,
				Flow:       true,
				Pressure:   false,
			},
			wantScore:  7, // 2+3+0+2+0
			wantLevel:  BuyPointStrong,
			wantAction: ActionBuyAll,
		},
		{
			name: "BUY - score >= 4",
			patterns: BuyPointPatterns{
				Pullback:   true,
				Breakout:   false,
				SecondWave: false,
				Flow:       true,
				Pressure:   true,
			},
			wantScore:  5, // 2+0+0+2+1
			wantLevel:  BuyPointNormal,
			wantAction: ActionAddPos,
		},
		{
			name: "BUY - borderline 4",
			patterns: BuyPointPatterns{
				Pullback:   true,
				Breakout:   false,
				SecondWave: false,
				Flow:       true,
				Pressure:   false,
			},
			wantScore:  4, // 2+0+0+2+0
			wantLevel:  BuyPointNormal,
			wantAction: ActionAddPos,
		},
		{
			name: "NO_BUY - no patterns",
			patterns: BuyPointPatterns{
				Pullback:   false,
				Breakout:   false,
				SecondWave: false,
				Flow:       false,
				Pressure:   false,
			},
			wantScore:  0,
			wantLevel:  BuyPointNone,
			wantAction: ActionWait,
		},
		{
			name: "NO_BUY - score < 4",
			patterns: BuyPointPatterns{
				Pullback:   true,
				Breakout:   false,
				SecondWave: false,
				Flow:       false,
				Pressure:   false,
			},
			wantScore:  2, // 2+0+0+0+0
			wantLevel:  BuyPointNone,
			wantAction: ActionWait,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, level, action := CalculateBuyPointScore(tt.patterns, nil)
			if score != tt.wantScore {
				t.Errorf("score: got %d, want %d", score, tt.wantScore)
			}
			if level != tt.wantLevel {
				t.Errorf("level: got %v, want %v", level, tt.wantLevel)
			}
			if action != tt.wantAction {
				t.Errorf("action: got %v, want %v", action, tt.wantAction)
			}
		})
	}
}

func TestAntiChase(t *testing.T) {
	tests := []struct {
		name  string
		bars  []interface{}
		want  bool
	}{
		{
			name:  "empty bars - no chase",
			bars:  []interface{}{},
			want:  false,
		},
		{
			name:  "single bar - no chase",
			bars:  []interface{}{BarPoint{Close: 100.0}},
			want:  false,
		},
		{
			name: "continuous big green - chase detected",
			bars: []interface{}{
				BarPoint{Close: 100.0},
				BarPoint{Close: 103.0}, // > 100 * 1.02 = 102
			},
			want: true,
		},
		{
			name: "continuous small green - no chase",
			bars: []interface{}{
				BarPoint{Close: 100.0},
				BarPoint{Close: 101.0}, // < 100 * 1.02
			},
			want: false,
		},
		{
			name: "drop then recover - chase detected (100.5 > 98*1.02)",
			bars: []interface{}{
				BarPoint{Close: 100.0},
				BarPoint{Close: 98.0},
				BarPoint{Close: 100.5},
			},
			want: true,
		},
		{
			name: "exact 2% - no chase",
			bars: []interface{}{
				BarPoint{Close: 100.0},
				BarPoint{Close: 102.0}, // exactly 1.02
			},
			want: false,
		},
		{
			name: "just above 2% - chase detected",
			bars: []interface{}{
				BarPoint{Close: 100.0},
				BarPoint{Close: 102.01},
			},
			want: true,
		},
		{
			name: "three consecutive big green",
			bars: []interface{}{
				BarPoint{Close: 100.0},
				BarPoint{Close: 103.0},
				BarPoint{Close: 106.0}, // > 103 * 1.02
			},
			want: true,
		},
		{
			name: "with BarPoint type",
			bars: []interface{}{
				BarPoint{Close: 100.0},
				BarPoint{Close: 103.0},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AntiChase(tt.bars)
			if got != tt.want {
				t.Errorf("AntiChase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBuyPointSignal(t *testing.T) {
	patterns := BuyPointPatterns{
		Pullback:   true,
		Breakout:   true,
		SecondWave: false,
		Flow:       true,
		Pressure:   true,
	}

	signal := NewBuyPointSignal("000001", patterns, false)

	if signal.Code != "000001" {
		t.Errorf("Code: got %s, want 000001", signal.Code)
	}
	if signal.Score != 8 { // 2+3+0+2+1
		t.Errorf("Score: got %d, want 8", signal.Score)
	}
	if signal.Signal != BuyPointStrong {
		t.Errorf("Signal: got %v, want BuyPointStrong", signal.Signal)
	}
	if signal.Action != ActionBuyAll {
		t.Errorf("Action: got %v, want ActionBuyAll", signal.Action)
	}
	if signal.AntiChase != false {
		t.Errorf("AntiChase: got %v, want false", signal.AntiChase)
	}
	if signal.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestNewBuyPointSignal_WithAntiChase(t *testing.T) {
	patterns := BuyPointPatterns{
		Pullback:   true,
		Breakout:   true,
		SecondWave: true,
		Flow:       true,
		Pressure:   true,
	}

	signal := NewBuyPointSignal("000001", patterns, true)

	if signal.AntiChase != true {
		t.Errorf("AntiChase: got %v, want true", signal.AntiChase)
	}
	// Even with high score, action should still reflect the anti_chase flag
	// Note: The current implementation doesn't change action based on anti_chase in the signal struct
}

func TestBuyPointSignal_String(t *testing.T) {
	patterns := BuyPointPatterns{
		Pullback:   true,
		Breakout:   true,
		SecondWave: false,
		Flow:       true,
		Pressure:   true,
	}
	signal := NewBuyPointSignal("000001", patterns, false)

	str := signal.String()
	if str == "" {
		t.Error("String() should not return empty")
	}
}

func TestBuyPointLevel_Constants(t *testing.T) {
	if BuyPointStrong != "STRONG_BUY" {
		t.Errorf("BuyPointStrong: got %s, want STRONG_BUY", BuyPointStrong)
	}
	if BuyPointNormal != "BUY" {
		t.Errorf("BuyPointNormal: got %s, want BUY", BuyPointNormal)
	}
	if BuyPointNone != "NO_BUY" {
		t.Errorf("BuyPointNone: got %s, want NO_BUY", BuyPointNone)
	}
}

func TestCheckPullback(t *testing.T) {
	tests := []struct {
		name         string
		currentPrice float64
		maPrice      float64
		want         bool
	}{
		{"within range 3%", 103.0, 100.0, true},
		{"at min 2%", 102.0, 100.0, true},
		{"at max 5%", 105.0, 100.0, true},
		{"below min", 101.0, 100.0, false},
		{"above max", 106.0, 100.0, false},
		{"zero ma", 103.0, 0.0, false},
		{"negative ma", 103.0, -1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPullback(tt.currentPrice, tt.maPrice)
			if got != tt.want {
				t.Errorf("CheckPullback() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckBreakout(t *testing.T) {
	tests := []struct {
		name       string
		curHigh    float64
		prevHigh   float64
		curVol     float64
		avgVol     float64
		want       bool
	}{
		{"breakout with volume", 105.0, 100.0, 200.0, 100.0, true},
		{"breakout without volume", 105.0, 100.0, 100.0, 100.0, false},
		{"no breakout with volume", 100.0, 105.0, 200.0, 100.0, false},
		{"exactly 1.5x volume", 105.0, 100.0, 150.0, 100.0, true},
		{"just below 1.5x", 105.0, 100.0, 149.0, 100.0, false},
		{"zero avg volume", 105.0, 100.0, 200.0, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckBreakout(tt.curHigh, tt.prevHigh, tt.curVol, tt.avgVol)
			if got != tt.want {
				t.Errorf("CheckBreakout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckSecondWave(t *testing.T) {
	tests := []struct {
		name       string
		curPrice   float64
		prevHigh   float64
		curVol     float64
		prevVol    float64
		want       bool
	}{
		{"valid second wave", 96.0, 100.0, 150.0, 100.0, true},
		{"near high but no vol", 96.0, 100.0, 90.0, 100.0, false},
		{"vol but not near high", 94.0, 100.0, 150.0, 100.0, false},
		{"exactly 95%", 95.0, 100.0, 150.0, 100.0, true},
		{"just below 95%", 94.9, 100.0, 150.0, 100.0, false},
		{"zero prev high", 96.0, 0.0, 150.0, 100.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckSecondWave(tt.curPrice, tt.prevHigh, tt.curVol, tt.prevVol)
			if got != tt.want {
				t.Errorf("CheckSecondWave() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckFlow(t *testing.T) {
	tests := []struct {
		name       string
		buyVol     float64
		sellVol    float64
		want       bool
	}{
		{"positive flow", 150.0, 100.0, true},
		{"neutral flow", 120.0, 100.0, false},
		{"negative flow", 100.0, 150.0, false},
		{"equal flow", 100.0, 100.0, false},
		{"zero sell", 100.0, 0.0, true},
		{"zero both", 0.0, 0.0, false},
		{"just above 1.2x", 121.0, 100.0, true},
		{"just below 1.2x", 119.0, 100.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckFlow(tt.buyVol, tt.sellVol)
			if got != tt.want {
				t.Errorf("CheckFlow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPressure(t *testing.T) {
	tests := []struct {
		name      string
		askVol    float64
		bidVol    float64
		want      bool
	}{
		{"no pressure", 80.0, 100.0, true},
		{"with pressure", 130.0, 100.0, false},
		{"exactly 1.2x", 120.0, 100.0, false},
		{"just above 1.2x", 121.0, 100.0, false},
		{"zero bid", 80.0, 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPressure(tt.askVol, tt.bidVol)
			if got != tt.want {
				t.Errorf("CheckPressure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBar(t *testing.T) {
	// Test Bar type
	bar := BarPoint{Open: 100, High: 105, Low: 99, Close: 103, Volume: 1000}
	result := getBar(bar)
	if result == nil || result.Close != 103 {
		t.Errorf("getBar(Bar): got %v, want Close=103", result)
	}

	// Test *Bar type
	barPtr := &BarPoint{Open: 100, High: 105, Low: 99, Close: 104, Volume: 1000}
	result = getBar(barPtr)
	if result == nil || result.Close != 104 {
		t.Errorf("getBar(*Bar): got %v, want Close=104", result)
	}

	// Test BarPoint type
	bp := BarPoint{Open: 100, High: 105, Low: 99, Close: 105, Volume: 1000}
	result = getBar(bp)
	if result == nil || result.Close != 105 {
		t.Errorf("getBar(BarPoint): got %v, want Close=105", result)
	}

	// Test *BarPoint type
	bpPtr := &BarPoint{Open: 100, High: 105, Low: 99, Close: 106, Volume: 1000}
	result = getBar(bpPtr)
	if result == nil || result.Close != 106 {
		t.Errorf("getBar(*BarPoint): got %v, want Close=106", result)
	}

	// Test map type
	mapBar := map[string]interface{}{
		"open":   100.0,
		"high":   105.0,
		"low":    99.0,
		"close":  107.0,
		"volume": 1000.0,
	}
	result = getBar(mapBar)
	if result == nil || result.Close != 107 {
		t.Errorf("getBar(map): got %v, want Close=107", result)
	}

	// Test nil
	result = getBar(nil)
	if result != nil {
		t.Errorf("getBar(nil): got %v, want nil", result)
	}
}

func TestDefaultBuyPointConfig(t *testing.T) {
	cfg := DefaultBuyPointConfig
	if cfg.PullbackMinPct != 2.0 {
		t.Errorf("PullbackMinPct: got %f, want 2.0", cfg.PullbackMinPct)
	}
	if cfg.PullbackMaxPct != 5.0 {
		t.Errorf("PullbackMaxPct: got %f, want 5.0", cfg.PullbackMaxPct)
	}
	if cfg.BreakoutVolMult != 1.5 {
		t.Errorf("BreakoutVolMult: got %f, want 1.5", cfg.BreakoutVolMult)
	}
	if cfg.SecondWaveThresh != 0.95 {
		t.Errorf("SecondWaveThresh: got %f, want 0.95", cfg.SecondWaveThresh)
	}
	if cfg.FlowRatio != 1.2 {
		t.Errorf("FlowRatio: got %f, want 1.2", cfg.FlowRatio)
	}
	if cfg.PressureRatio != 1.2 {
		t.Errorf("PressureRatio: got %f, want 1.2", cfg.PressureRatio)
	}
}

func TestBuyPointSignal_Timestamp(t *testing.T) {
	before := time.Now()
	patterns := BuyPointPatterns{}
	signal := NewBuyPointSignal("000001", patterns, false)
	after := time.Now()

	if signal.Timestamp.Before(before) || signal.Timestamp.After(after) {
		t.Error("Timestamp should be between before and after creation")
	}
}

func TestCalculateBuyPointScore_WithConfig(t *testing.T) {
	patterns := BuyPointPatterns{
		Pullback:   true,
		Breakout:   true,
		SecondWave: false,
		Flow:       true,
		Pressure:   true,
	}

	// Use custom config
	config := &BuyPointConfig{
		PullbackMinPct:   1.0,
		PullbackMaxPct:   10.0,
		BreakoutVolMult:  2.0,
		SecondWaveThresh: 0.90,
		FlowRatio:        1.5,
		PressureRatio:    1.5,
	}

	score, level, action := CalculateBuyPointScore(patterns, config)
	// With nil config it should be same result
	scoreNil, levelNil, actionNil := CalculateBuyPointScore(patterns, nil)

	if score != scoreNil || level != levelNil || action != actionNil {
		t.Errorf("Custom config should not affect score calculation in this test")
	}
}
