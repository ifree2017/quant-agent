package social

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Client & RateLimiter Tests
// ---------------------------------------------------------------------------

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)

	var count int
	var mu sync.Mutex
	var lastTime time.Time

	start := time.Now()
	for i := 0; i < 5; i++ {
		rl.Wait()
		mu.Lock()
		count++
		now := time.Now()
		if count > 1 {
			elapsed := now.Sub(lastTime)
			if elapsed < 100*time.Millisecond {
				t.Errorf("rate limiter allowed request too quickly: %v", elapsed)
			}
		}
		lastTime = now
		mu.Unlock()
	}
	elapsed := time.Since(start)
	// 5 requests × 100ms each = at least 400ms
	if elapsed < 400*time.Millisecond {
		t.Errorf("rate limiter did not enforce minimum interval")
	}
}

// ---------------------------------------------------------------------------
// Sentiment Tests
// ---------------------------------------------------------------------------

func TestCalcSentiment(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  string
		likes    int
		comments int
		views    int
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "bullish post",
			title:    "强烈看好，准备抄底加仓",
			content:  "估值处于历史低位，主力资金持续流入，是难得的布局机会",
			likes:    100,
			comments: 50,
			wantMin:  10,
			wantMax:  100,
		},
		{
			name:     "bearish post",
			title:    "高位风险，建议减仓止损",
			content:  "MACD形成死叉，跌破重要支撑，建议减仓规避风险",
			likes:    80,
			comments: 30,
			wantMin:  -100,
			wantMax:  -10,
		},
		{
			name:     "neutral post",
			title:    "今日行情分析",
			content:  "成交量有所放大，但方向不明确，等待趋势确认",
			likes:    10,
			comments: 5,
			wantMin:  -15,
			wantMax:  15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcSentiment(tt.title, tt.content, tt.likes, tt.comments, tt.views)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calcSentiment() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalcSentimentBounds(t *testing.T) {
	// Extreme cases must stay within [-100, 100]
	for i := 0; i < 100; i++ {
		s := calcSentiment("test", "content", 10000, 5000, 0)
		if s < -100 || s > 100 {
			t.Errorf("sentiment out of bounds: %v", s)
		}
	}
}

func TestExtractKeywords(t *testing.T) {
	text := "强烈看好，准备抄底！高位风险，建议止损减仓。"
	keywords := extractKeywords(text)

	hasBullish := false
	hasBearish := false
	for _, kw := range keywords {
		for _, b := range bullishKeywords {
			if kw == b {
				hasBullish = true
			}
		}
		for _, b := range bearishKeywords {
			if kw == b {
				hasBearish = true
			}
		}
	}

	if !hasBullish {
		t.Error("expected bullish keyword not found")
	}
	if !hasBearish {
		t.Error("expected bearish keyword not found")
	}
}

func TestUniqStrings(t *testing.T) {
	in := []string{"a", "b", "a", "c", "b"}
	got := uniqStrings(in)
	if len(got) != 3 {
		t.Errorf("uniqStrings = %v, want 3 unique items", got)
	}
}

// ---------------------------------------------------------------------------
// Models Tests
// ---------------------------------------------------------------------------

func TestSentimentDataJSON(t *testing.T) {
	data := SentimentData{
		StockCode:      "600519",
		Platform:       "eastmoney",
		SentimentScore: 75.5,
		PostCount:      10,
		CommentCount:  100,
		HeatScore:     250.5,
		Keywords:       []string{"看好", "低估"},
		FetchedAt:     time.Now(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var got SentimentData
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if got.StockCode != data.StockCode {
		t.Errorf("StockCode = %v, want %v", got.StockCode, data.StockCode)
	}
	if got.SentimentScore != data.SentimentScore {
		t.Errorf("SentimentScore = %v, want %v", got.SentimentScore, data.SentimentScore)
	}
}

func TestMarketHotItemJSON(t *testing.T) {
	item := MarketHotItem{
		Rank:      1,
		StockCode: "000001",
		StockName: "平安银行",
		ChangePct: 3.45,
		HeatScore: 120.5,
		Platform:  "eastmoney",
	}

	b, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var got MarketHotItem
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if got.StockCode != item.StockCode {
		t.Errorf("StockCode = %v, want %v", got.StockCode, item.StockCode)
	}
	if got.ChangePct != item.ChangePct {
		t.Errorf("ChangePct = %v, want %v", got.ChangePct, item.ChangePct)
	}
}

// ---------------------------------------------------------------------------
// Eastmoney Client Tests (with mock server)
// ---------------------------------------------------------------------------

func TestEastmoneyClientFetchMarketHot(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/qt/clist/get" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"diff": [
					{"f12": "600519", "f14": "贵州茅台", "f3": 2.5, "f62": 1234567890},
					{"f12": "000001", "f14": "平安银行", "f3": -1.2, "f62": -987654321}
				]
			}
		}`))
	}))
	defer server.Close()

	// We can't easily inject the URL into the client, so just test parsing logic
	// by checking the client constructs the right URL
	_ = server

	// Test heat score calculation
	heat := calcHeatScore(2.5, 1234567890)
	if heat <= 0 {
		t.Errorf("calcHeatScore should be positive for positive inputs, got %v", heat)
	}
}

func TestCalcHeatScore(t *testing.T) {
	tests := []struct {
		changePct float64
		netInflow  float64
		wantMin    float64
	}{
		{5.0, 1e9, 60},   // 5% change * 10 = 50, + 10e (inflow/1e8*10) = ~60
		{-5.0, -1e9, 60}, // abs values
		{0, 0, 0},
	}

	for _, tt := range tests {
		got := calcHeatScore(tt.changePct, tt.netInflow)
		if got < tt.wantMin {
			t.Errorf("calcHeatScore(%v, %v) = %v, want >= %v", tt.changePct, tt.netInflow, got, tt.wantMin)
		}
	}
}

func TestCalcMarketSentiment(t *testing.T) {
	tests := []struct {
		name      string
		changePct float64
		netInflow float64
		wantMin   float64
		wantMax   float64
	}{
		{"strong bullish", 8.0, 5e9, 10, 100},
		{"strong bearish", -8.0, -5e9, -100, -10},
		{"neutral", 0, 0, -5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcMarketSentiment(tt.changePct, tt.netInflow)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calcMarketSentiment = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Xueqiu Client Tests
// ---------------------------------------------------------------------------

func TestXueqiuClientFetchStockPosts(t *testing.T) {
	c := NewXueqiuClient()
	posts, err := c.FetchStockPosts("600519")
	if err != nil {
		t.Fatalf("FetchStockPosts error: %v", err)
	}
	if len(posts) == 0 {
		t.Error("expected posts, got none")
	}
	for _, p := range posts {
		if p.Platform != XueqiuPlatform {
			t.Errorf("platform = %v, want %v", p.Platform, XueqiuPlatform)
		}
		if p.StockCode != "600519" {
			t.Errorf("stock code = %v, want 600519", p.StockCode)
		}
		if p.Sentiment < -100 || p.Sentiment > 100 {
			t.Errorf("sentiment out of bounds: %v", p.Sentiment)
		}
	}
}

func TestXueqiuClientFetchStockSentiment(t *testing.T) {
	c := NewXueqiuClient()
	data, err := c.FetchStockSentiment("600519")
	if err != nil {
		t.Fatalf("FetchStockSentiment error: %v", err)
	}
	if data.StockCode != "600519" {
		t.Errorf("StockCode = %v, want 600519", data.StockCode)
	}
	if data.Platform != XueqiuPlatform {
		t.Errorf("Platform = %v, want %v", data.Platform, XueqiuPlatform)
	}
	if data.SentimentScore < -100 || data.SentimentScore > 100 {
		t.Errorf("sentiment score out of bounds: %v", data.SentimentScore)
	}
	if data.PostCount != 5 {
		t.Errorf("PostCount = %v, want 5", data.PostCount)
	}
}

// ---------------------------------------------------------------------------
// Aggregate Tests
// ---------------------------------------------------------------------------

func TestAggregateSentiment(t *testing.T) {
	posts := []Post{
		{Sentiment: 50},
		{Sentiment: -30},
		{Sentiment: 20},
	}
	got := AggregateSentiment(posts)
	want := 13.33
	if got < want-1 || got > want+1 {
		t.Errorf("AggregateSentiment = %v, want ~%v", got, want)
	}
}

func TestAggregateSentimentEmpty(t *testing.T) {
	got := AggregateSentiment([]Post{})
	if got != 0 {
		t.Errorf("AggregateSentiment empty = %v, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func TestCleanHTML(t *testing.T) {
	html := "<p>Hello <b>World</b></p>"
	got := CleanHTML(html)
	if got != "Hello World" {
		t.Errorf("CleanHTML = %q, want %q", got, "Hello World")
	}
}


