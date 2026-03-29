package signal

import (
	"testing"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("NewEngine should not return nil")
	}
	if engine.socialClient == nil {
		t.Error("socialClient should not be nil")
	}
}

func TestApplyRules(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name       string
		score      float64
		confidence float64
		want       SignalType
	}{
		// BUY: score > 60 && confidence > 0.7 (strictly greater than per spec)
		{"strong buy", 75, 0.8, SignalBuy},
		{"buy at threshold (60)", 60, 0.7, SignalWatch},    // 60 not > 60, but passes WATCH
		{"buy low confidence", 75, 0.6, SignalWatch},       // 0.6 not > 0.7 for BUY, but passes WATCH

		// SELL: score < -60 && confidence > 0.7 (strictly less than per spec)
		{"strong sell", -75, 0.8, SignalSell},
		{"sell at threshold (-60)", -60, 0.7, SignalHold},  // -60 not < -60, score not > 30 either
		{"sell low confidence", -75, 0.5, SignalHold},       // 0.5 not > 0.7, but -75 < -30 so HOLD

		// WATCH: score > 30 && confidence > 0.5 (strictly greater than per spec)
		{"watch", 50, 0.6, SignalWatch},
		{"watch at threshold (30)", 30, 0.5, SignalHold},  // 30 not > 30

		// HOLD: -30 <= score <= 30
		{"hold neutral", 0, 0.8, SignalHold},
		{"hold positive bound", 30, 0.8, SignalHold},
		{"hold negative bound", -30, 0.8, SignalHold},
		{"hold low confidence high score", 65, 0.4, SignalHold}, // 65 > 30 but 0.4 not > 0.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := engine.applyRules(tt.score, tt.confidence)
			if got != tt.want {
				t.Errorf("applyRules(%v, %v): got %s, want %s", tt.score, tt.confidence, got, tt.want)
			}
		})
	}
}

func TestGenerateMessage(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		score    float64
		keywords []string
		stock    string
		checkFn  func(string) bool
	}{
		{"strong bullish", 75, []string{"看好", "低估"}, "000001", func(msg string) bool {
			return containsString(msg, "强烈看多") && containsString(msg, "75")
		}},
		{"bearish", -70, []string{"减持", "风险"}, "000002", func(msg string) bool {
			return containsString(msg, "强烈看空") && containsString(msg, "-70")
		}},
		{"neutral", 10, []string{}, "000003", func(msg string) bool {
			return containsString(msg, "略偏多") && containsString(msg, "10")
		}},
		{"with keywords", 55, []string{"看好", "低估", "龙头"}, "000004", func(msg string) bool {
			return containsString(msg, "关键词")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := engine.generateMessage(tt.score, tt.keywords, tt.stock)
			if !tt.checkFn(msg) {
				t.Errorf("generateMessage(%v, %v, %v): got %s, check failed", tt.score, tt.keywords, tt.stock, msg)
			}
		})
	}
}

func TestMatchSectors(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{"新能源 keywords", "宁德时代 新能源 锂电", []string{"新能源"}},
		{"半导体 keywords", "芯片 半导体 集成电路", []string{"半导体"}},
		{"多板块", "AI 人工智能 芯片", []string{"人工智能", "半导体"}}, // order may vary
		{"no match", "日常生活 吃饭睡觉", []string{}},
		{"partial match", "光伏 太阳能", []string{"新能源"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.matchSectors(tt.text)
			// 使用集合比较（忽略顺序）
			gotSet := make(map[string]bool)
			for _, v := range got {
				gotSet[v] = true
			}
			expectedSet := make(map[string]bool)
			for _, v := range tt.expected {
				expectedSet[v] = true
			}
			if len(gotSet) != len(expectedSet) {
				t.Errorf("matchSectors(%s): got %v, want %v", tt.text, got, tt.expected)
				return
			}
			for k := range gotSet {
				if !expectedSet[k] {
					t.Errorf("matchSectors(%s): unexpected element %s", tt.text, k)
				}
			}
		})
	}
}

func TestExtractHotKeywords(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		text     string
		minCount int
	}{
		{"simple text", "苹果公司 财报", 1},
		{"hashtag style", "新能源 宁德时代", 1}, // with space separator
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.extractHotKeywords(tt.text)
			if tt.minCount > 0 && len(got) < tt.minCount {
				t.Errorf("extractHotKeywords(%s): got %d keywords, want at least %d", tt.text, len(got), tt.minCount)
			}
		})
	}
}

func TestDetectDirection(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name   string
		signal *SectorSignal
		want   string
	}{
		{"strong emerging", &SectorSignal{ChangePct: 8}, "emerging"},
		{"mild emerging", &SectorSignal{ChangePct: 3}, "stable"},
		{"fading", &SectorSignal{ChangePct: -5}, "fading"},
		{"stable", &SectorSignal{ChangePct: 1}, "stable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.detectDirection(tt.signal)
			if got != tt.want {
				t.Errorf("detectDirection: got %s, want %s", got, tt.want)
			}
		})
	}
}

// helper functions

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
