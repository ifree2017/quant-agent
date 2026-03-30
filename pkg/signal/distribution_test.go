package signal

import (
	"testing"
	"time"
)

func TestDistributionFeatures_ScoreCalculation(t *testing.T) {
	tests := []struct {
		name     string
		features DistributionFeatures
		wantScore int
		wantLevel DistributionLevel
		wantAction TradeAction
	}{
		{
			name: "STRONG_DISTRIBUTION - all features positive",
			features: DistributionFeatures{
				SpikeFall:       true,
				SecondWave:      false,
				VolumeDecay:     true,
				SellPressure:    true,
				ActiveSellRatio: 0.5,
			},
			wantScore:  9, // 2+2+1+2+2
			wantLevel:  DistributionStrong,
			wantAction: ActionSellAll,
		},
		{
			name: "STRONG_DISTRIBUTION - score >= 7",
			features: DistributionFeatures{
				SpikeFall:       true,
				SecondWave:      false,
				VolumeDecay:     true,
				SellPressure:    true,
				ActiveSellRatio: 0.9, // >= 0.8, no +2
			},
			wantScore:  7, // 2+2+1+2+0
			wantLevel:  DistributionStrong,
			wantAction: ActionSellAll,
		},
		{
			name: "DISTRIBUTION - score = 5",
			features: DistributionFeatures{
				SpikeFall:       true,
				SecondWave:      true, // second wave present, no +2
				VolumeDecay:     true,
				SellPressure:    true,
				ActiveSellRatio: 0.9,
			},
			wantScore:  5, // 2+0+1+2+0 = 5
			wantLevel:  DistributionNormal,
			wantAction: ActionReduce,
		},
		{
			name: "DISTRIBUTION - borderline 4",
			features: DistributionFeatures{
				SpikeFall:       true,
				SecondWave:      true, // has second wave, no +2
				VolumeDecay:     false,
				SellPressure:    true,
				ActiveSellRatio: 0.9,
			},
			wantScore:  4, // 2+0+0+2+0 = 4
			wantLevel:  DistributionNormal,
			wantAction: ActionReduce,
		},
		{
			name: "STRONG_DISTRIBUTION - borderline 7 (score=6)",
			features: DistributionFeatures{
				SpikeFall:       true,
				SecondWave:      false, // +2
				VolumeDecay:     false,
				SellPressure:    true,  // +2
				ActiveSellRatio: 0.9,   // +0
			},
			wantScore:  6, // 2+2+0+2+0 = 6
			wantLevel:  DistributionNormal,
			wantAction: ActionReduce,
		},
		{
			name: "NEUTRAL - no features",
			features: DistributionFeatures{
				SpikeFall:       false,
				SecondWave:      true,
				VolumeDecay:     false,
				SellPressure:    false,
				ActiveSellRatio: 1.2,
			},
			wantScore:  0,
			wantLevel:  DistributionNeutral,
			wantAction: ActionHold,
		},
		{
			name: "NEUTRAL - partial features score=3",
			features: DistributionFeatures{
				SpikeFall:       true,
				SecondWave:      true,
				VolumeDecay:     true,
				SellPressure:    false,
				ActiveSellRatio: 1.2,
			},
			wantScore:  3, // 2+0+1+0+0
			wantLevel:  DistributionNeutral,
			wantAction: ActionHold,
		},
		{
			name: "Edge case - ActiveSellRatio exactly 0.8",
			features: DistributionFeatures{
				SpikeFall:       false,
				SecondWave:      false,
				VolumeDecay:     false,
				SellPressure:    false,
				ActiveSellRatio: 0.8,
			},
			wantScore:  2, // 0+2+0+0+0
			wantLevel:  DistributionNeutral,
			wantAction: ActionHold,
		},
		{
			name: "Edge case - ActiveSellRatio just below 0.8",
			features: DistributionFeatures{
				SpikeFall:       false,
				SecondWave:      false,
				VolumeDecay:     false,
				SellPressure:    false,
				ActiveSellRatio: 0.79,
			},
			wantScore:  4, // 0+2+0+0+2
			wantLevel:  DistributionNormal,
			wantAction: ActionReduce,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, level, action := CalculateDistributionScore(tt.features)
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

func TestNewDistributionSignal(t *testing.T) {
	features := DistributionFeatures{
		SpikeFall:       true,
		SecondWave:      false,
		VolumeDecay:     true,
		SellPressure:    true,
		ActiveSellRatio: 0.5,
	}

	signal := NewDistributionSignal("000001", features)

	if signal.Code != "000001" {
		t.Errorf("Code: got %s, want 000001", signal.Code)
	}
	if signal.Score != 9 {
		t.Errorf("Score: got %d, want 9", signal.Score)
	}
	if signal.Signal != DistributionStrong {
		t.Errorf("Signal: got %v, want DistributionStrong", signal.Signal)
	}
	if signal.Action != ActionSellAll {
		t.Errorf("Action: got %v, want ActionSellAll", signal.Action)
	}
	if signal.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestDistributionSignal_String(t *testing.T) {
	features := DistributionFeatures{
		SpikeFall:       true,
		SecondWave:      false,
		VolumeDecay:     true,
		SellPressure:    true,
		ActiveSellRatio: 0.5,
	}
	signal := NewDistributionSignal("000001", features)

	str := signal.String()
	if str == "" {
		t.Error("String() should not return empty")
	}
}

func TestDistributionLevel_Constants(t *testing.T) {
	if DistributionStrong != "STRONG_DISTRIBUTION" {
		t.Errorf("DistributionStrong: got %s, want STRONG_DISTRIBUTION", DistributionStrong)
	}
	if DistributionNormal != "DISTRIBUTION" {
		t.Errorf("DistributionNormal: got %s, want DISTRIBUTION", DistributionNormal)
	}
	if DistributionNeutral != "NEUTRAL" {
		t.Errorf("DistributionNeutral: got %s, want NEUTRAL", DistributionNeutral)
	}
}

func TestTradeAction_Constants(t *testing.T) {
	if ActionSellAll != "SELL_ALL" {
		t.Errorf("ActionSellAll: got %s, want SELL_ALL", ActionSellAll)
	}
	if ActionReduce != "REDUCE" {
		t.Errorf("ActionReduce: got %s, want REDUCE", ActionReduce)
	}
	if ActionHold != "HOLD" {
		t.Errorf("ActionHold: got %s, want HOLD", ActionHold)
	}
	if ActionBuyAll != "BUY_ALL" {
		t.Errorf("ActionBuyAll: got %s, want BUY_ALL", ActionBuyAll)
	}
	if ActionAddPos != "ADD_POSITION" {
		t.Errorf("ActionAddPos: got %s, want ADD_POSITION", ActionAddPos)
	}
	if ActionWait != "WAIT" {
		t.Errorf("ActionWait: got %s, want WAIT", ActionWait)
	}
}

func TestDistributionFeatures_AllCombinations(t *testing.T) {
	// Test score boundaries
	testCases := []struct {
		spikeFall       bool
		secondWave      bool
		volumeDecay     bool
		sellPressure    bool
		activeSellRatio float64
		minScore        int
		maxScore        int
	}{
		{true, false, true, true, 0.5, 7, 9},
		{false, false, false, false, 1.5, 2, 2},
		{true, true, true, true, 0.5, 5, 7},
	}

	for i, tc := range testCases {
		features := DistributionFeatures{
			SpikeFall:       tc.spikeFall,
			SecondWave:      tc.secondWave,
			VolumeDecay:     tc.volumeDecay,
			SellPressure:    tc.sellPressure,
			ActiveSellRatio: tc.activeSellRatio,
		}
		score, _, _ := CalculateDistributionScore(features)
		if score < tc.minScore || score > tc.maxScore {
			t.Errorf("case %d: score %d not in range [%d, %d]", i, score, tc.minScore, tc.maxScore)
		}
	}
}

func TestDistributionSignal_Timestamp(t *testing.T) {
	before := time.Now()
	features := DistributionFeatures{}
	signal := NewDistributionSignal("000001", features)
	after := time.Now()

	if signal.Timestamp.Before(before) || signal.Timestamp.After(after) {
		t.Error("Timestamp should be between before and after creation")
	}
}
