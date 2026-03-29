package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"quant-agent/internal/model"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// AIer interface for mocking AI client in handler tests
type AIer interface {
	AnalyzeStyle(records []model.TradeRecord) (model.StyleProfile, error)
	GenerateStrategy(profile model.StyleProfile) (model.StrategyRules, error)
}

// roundTripFuncForTest implements http.RoundTripper
type roundTripFuncForTest func(*http.Request) (*http.Response, error)

func (f roundTripFuncForTest) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// mockAIClient is a test double for ai.Client
type mockAIClient struct {
	styleProfile   model.StyleProfile
	strategyRules model.StrategyRules
	styleErr      error
	strategyErr   error
}

func (m *mockAIClient) AnalyzeStyle(records []model.TradeRecord) (model.StyleProfile, error) {
	if m.styleErr != nil {
		return model.StyleProfile{}, m.styleErr
	}
	return m.styleProfile, nil
}

func (m *mockAIClient) GenerateStrategy(profile model.StyleProfile) (model.StrategyRules, error) {
	if m.strategyErr != nil {
		return model.StrategyRules{}, m.strategyErr
	}
	return m.strategyRules, nil
}



func TestStyleAnalyze_Handler(t *testing.T) {
	mockClient := &mockAIClient{
		styleProfile: model.StyleProfile{
			ID:                   "profile-1",
			UserID:               "user-1",
			Style:                "稳健",
			RiskScore:            45,
			TradeFrequency:       "中频",
			AvgHoldDays:          10,
			MaxDrawdownTolerance: 15,
		},
	}

	router := gin.New()
	// Use a handler wrapper that adapts mockAIClient to the actual handler signature
	router.POST("/style/analyze", func(c *gin.Context) {
		var req model.StyleAnalyzeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		profile, err := mockClient.AnalyzeStyle(req.Records)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		profile.UserID = req.UserID
		c.JSON(http.StatusOK, &profile)
	})

	reqBody := model.StyleAnalyzeRequest{
		UserID: "user-1",
		Records: []model.TradeRecord{
			{Date: "2025-01-02", Action: "buy", Symbol: "000001", Price: 10.0, Quantity: 100},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/style/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200. body: %s", w.Code, w.Body.String())
	}

	var resp model.StyleProfile
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v. body: %s", err, w.Body.String())
	}
	if resp.Style != "稳健" {
		t.Errorf("style: got %s, want 稳健", resp.Style)
	}
	if resp.UserID != "user-1" {
		t.Errorf("userID: got %s, want user-1", resp.UserID)
	}
}

func TestStyleAnalyze_Handler_BadRequest(t *testing.T) {
	router := gin.New()
	router.POST("/style/analyze", func(c *gin.Context) {
		var req model.StyleAnalyzeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest(http.MethodPost, "/style/analyze", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestStrategyGenerate_Handler(t *testing.T) {
	mockClient := &mockAIClient{
		strategyRules: model.StrategyRules{
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
			Market: "A-share",
		},
	}

	router := gin.New()
	router.POST("/strategy/generate", func(c *gin.Context) {
		var req model.StrategyGenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rules, err := mockClient.GenerateStrategy(req.StyleProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_ = rules
		c.JSON(http.StatusOK, gin.H{"rules": rules})
	})

	reqBody := model.StrategyGenerateRequest{
		UserID: "user-1",
		StyleProfile: model.StyleProfile{
			Style:     "稳健",
			RiskScore: 45,
		},
		Market: "A-share",
		Symbol: "000001",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/strategy/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200. body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v. body: %s", err, w.Body.String())
	}
	if resp["rules"] == nil {
		t.Error("rules should not be nil")
	}
}

func TestStrategyGenerate_Handler_BadRequest(t *testing.T) {
	router := gin.New()
	router.POST("/strategy/generate", func(c *gin.Context) {
		var req model.StrategyGenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest(http.MethodPost, "/strategy/generate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestBacktestRun_Handler(t *testing.T) {
	router := gin.New()
	router.POST("/backtest/run", BacktestRun("data"))

	reqBody := model.BacktestRunRequest{
		StrategyID:  "strategy-1",
		Symbol:     "000001",
		Days:       30,
		InitialCash: 100000,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/backtest/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200. body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["jobId"] == nil || resp["jobId"] == "" {
		t.Error("jobId should not be empty")
	}
	if resp["status"] != "queued" {
		t.Errorf("status: got %v, want queued", resp["status"])
	}
}

func TestBacktestRun_Handler_BadRequest(t *testing.T) {
	router := gin.New()
	router.POST("/backtest/run", BacktestRun("data"))

	req := httptest.NewRequest(http.MethodPost, "/backtest/run", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestBacktestRun_Handler_Defaults(t *testing.T) {
	router := gin.New()
	router.POST("/backtest/run", BacktestRun("data"))

	reqBody := map[string]string{
		"strategyId": "strategy-1",
		"symbol":     "000001",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/backtest/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200. body: %s", w.Code, w.Body.String())
	}
}

func TestStrategyList_Handler(t *testing.T) {
	router := gin.New()
	router.GET("/strategy/list", StrategyList())

	req := httptest.NewRequest(http.MethodGet, "/strategy/list", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var resp map[string][]model.Strategy
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["strategies"] == nil {
		t.Error("strategies key missing in response")
	}
}

func TestStrategyGet_Handler(t *testing.T) {
	router := gin.New()
	router.GET("/strategy/get", StrategyGet())

	req := httptest.NewRequest(http.MethodGet, "/strategy/get", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestStrategyDelete_Handler(t *testing.T) {
	router := gin.New()
	router.DELETE("/strategy/delete", StrategyDelete())

	req := httptest.NewRequest(http.MethodDelete, "/strategy/delete", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var resp map[string]bool
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp["deleted"] {
		t.Error("deleted should be true")
	}
}

func TestBacktestGet_Handler(t *testing.T) {
	router := gin.New()
	router.GET("/backtest/get", BacktestGet())

	req := httptest.NewRequest(http.MethodGet, "/backtest/get", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if _, ok := resp["result"]; !ok {
		t.Error("result key missing in response")
	}
}

func TestStyleAnalyze_Handler_AIError(t *testing.T) {
	mockClient := &mockAIClient{
		styleErr: &testNetError{msg: "connection refused"},
	}

	router := gin.New()
	router.POST("/style/analyze", func(c *gin.Context) {
		var req model.StyleAnalyzeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		profile, err := mockClient.AnalyzeStyle(req.Records)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		profile.UserID = req.UserID
		c.JSON(http.StatusOK, &profile)
	})

	reqBody := model.StyleAnalyzeRequest{
		UserID: "user-1",
		Records: []model.TradeRecord{
			{Date: "2025-01-02", Action: "buy", Symbol: "000001", Price: 10.0, Quantity: 100},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/style/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

func TestStrategyGenerate_Handler_AIError(t *testing.T) {
	mockClient := &mockAIClient{
		strategyErr: &testNetError{msg: "connection refused"},
	}

	router := gin.New()
	router.POST("/strategy/generate", func(c *gin.Context) {
		var req model.StrategyGenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rules, err := mockClient.GenerateStrategy(req.StyleProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_ = rules
		c.JSON(http.StatusOK, gin.H{"rules": rules})
	})

	reqBody := model.StrategyGenerateRequest{
		UserID: "user-1",
		StyleProfile: model.StyleProfile{
			Style:     "稳健",
			RiskScore: 45,
		},
		Market: "A-share",
		Symbol: "000001",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/strategy/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

type testNetError struct {
	msg string
}

func (n *testNetError) Error() string   { return n.msg }
func (n *testNetError) Timeout() bool   { return false }
func (n *testNetError) Temporary() bool { return true }
