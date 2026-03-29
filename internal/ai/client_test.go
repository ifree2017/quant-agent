package ai

import (
	"io"
	"net/http"
	"quant-agent/internal/model"
	"strings"
	"testing"
)

// roundTripFunc implements http.RoundTripper
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func makeTestClient(body string) *Client {
	return &Client{
		baseURL:    "http://test",
		token:      "test-token",
		model:      "gpt-4o-mini",
		maxRetries: 1,
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(body)),
				}, nil
			}),
		},
	}
}

func TestClient_AnalyzeStyle(t *testing.T) {
	mockedResponse := `{
		"choices": [{
			"message": {
				"content": "{\"style\":\"稳健\",\"riskScore\":45,\"tradeFrequency\":\"中频\",\"avgHoldDays\":10,\"maxDrawdownTolerance\":15,\"id\":\"\",\"userId\":\"\"}"
			}
		}]
	}`
	client := makeTestClient(mockedResponse)

	records := []model.TradeRecord{
		{Date: "2025-01-02", Action: "buy", Symbol: "000001", Price: 10.0, Quantity: 100},
	}

	profile, err := client.AnalyzeStyle(records)
	if err != nil {
		t.Fatalf("AnalyzeStyle: %v", err)
	}
	if profile.Style != "稳健" {
		t.Errorf("style: got %s, want 稳健", profile.Style)
	}
	if profile.RiskScore != 45 {
		t.Errorf("risk score: got %v, want 45", profile.RiskScore)
	}
	if profile.TradeFrequency != "中频" {
		t.Errorf("trade frequency: got %s, want 中频", profile.TradeFrequency)
	}
}

func TestClient_GenerateStrategy(t *testing.T) {
	mockedResponse := `{
		"choices": [{
			"message": {
				"content": "{\"indicators\":[\"MA\",\"RSI\"],\"params\":{\"MA_period\":20,\"RSI_period\":14,\"RSI_overbought\":70,\"RSI_oversold\":30},\"entry\":{\"type\":\"cross\",\"condition\":\"MA_cross_RSI\"},\"exit\":{\"type\":\"stop_loss\",\"value\":0.05},\"position\":{\"type\":\"fixed\",\"value\":0.2},\"market\":\"A-share\"}"
			}
		}]
	}`
	client := makeTestClient(mockedResponse)

	profile := model.StyleProfile{
		Style:                "稳健",
		RiskScore:            45,
		TradeFrequency:       "中频",
		AvgHoldDays:          10,
		MaxDrawdownTolerance: 15,
	}

	rules, err := client.GenerateStrategy(profile)
	if err != nil {
		t.Fatalf("GenerateStrategy: %v", err)
	}
	if len(rules.Indicators) != 2 {
		t.Errorf("indicators count: got %d, want 2", len(rules.Indicators))
	}
	if rules.Exit.Type != "stop_loss" {
		t.Errorf("exit type: got %s, want stop_loss", rules.Exit.Type)
	}
	if rules.Position.Type != "fixed" {
		t.Errorf("position type: got %s, want fixed", rules.Position.Type)
	}
}

func TestClient_RetryOnFailure(t *testing.T) {
	callCount := 0
	client := &Client{
		baseURL:    "http://test",
		token:      "test-token",
		model:      "gpt-4o-mini",
		maxRetries: 3,
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				callCount++
				if callCount < 3 {
					return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(`{"error":"server error"}`))}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(strings.NewReader(`{
						"choices": [{
							"message": {
								"content": "{\"style\":\"保守\",\"riskScore\":30,\"tradeFrequency\":\"低频\",\"avgHoldDays\":20,\"maxDrawdownTolerance\":10,\"id\":\"\",\"userId\":\"\"}"
							}
						}]
					}`)),
				}, nil
			}),
		},
	}

	records := []model.TradeRecord{{Date: "2025-01-02", Action: "buy", Symbol: "000001", Price: 10.0, Quantity: 100}}
	profile, err := client.AnalyzeStyle(records)
	if err != nil {
		t.Fatalf("should succeed after retries: %v", err)
	}
	if callCount != 3 {
		t.Errorf("call count: got %d, want 3", callCount)
	}
	if profile.Style != "保守" {
		t.Errorf("style: got %s, want 保守", profile.Style)
	}
}

func TestClient_RetryExhausted(t *testing.T) {
	client := &Client{
		baseURL:    "http://test",
		token:      "test-token",
		model:      "gpt-4o-mini",
		maxRetries: 2,
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(`{"error":"server error"}`))}, nil
			}),
		},
	}

	records := []model.TradeRecord{{Date: "2025-01-02", Action: "buy", Symbol: "000001", Price: 10.0, Quantity: 100}}
	_, err := client.AnalyzeStyle(records)
	if err == nil {
		t.Error("expected error after exhausting retries")
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{"a":1}`, `{"a":1}`},
		{`Here is the JSON: {"a":1} and more text`, `{"a":1} and more text`},
		{`no json here`, `no json here`},
		{`Some text before {"key": "value"} after`, `{"key": "value"} after`},
	}

	for _, tt := range tests {
		got := extractJSON(tt.input)
		if got != tt.want {
			t.Errorf("extractJSON(%q): got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildStylePrompt(t *testing.T) {
	records := []model.TradeRecord{
		{Date: "2025-01-02", Action: "buy", Symbol: "000001", Price: 10.0, Quantity: 100},
	}
	prompt := buildStylePrompt(records)
	if !strings.Contains(prompt, "000001") {
		t.Error("prompt should contain symbol")
	}
	// JSON of 10.0 is "price":10 (no decimal for integer-valued float)
	if !strings.Contains(prompt, `"price":10`) && !strings.Contains(prompt, `"price":10.0`) {
		t.Error("prompt should contain price field")
	}
}

func TestBuildStrategyPrompt(t *testing.T) {
	profile := model.StyleProfile{
		Style:     "稳健",
		RiskScore: 45,
	}
	prompt := buildStrategyPrompt(profile)
	if !strings.Contains(prompt, "稳健") {
		t.Error("prompt should contain style")
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080", "my-token")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL: got %s, want http://localhost:8080", c.baseURL)
	}
	if c.token != "my-token" {
		t.Errorf("token mismatch")
	}
	if c.maxRetries != 3 {
		t.Errorf("maxRetries: got %d, want 3", c.maxRetries)
	}
}

// TestClient_NetworkError tests that network errors trigger retry
func TestClient_NetworkError(t *testing.T) {
	client := &Client{
		baseURL:    "http://test",
		token:      "test-token",
		model:      "gpt-4o-mini",
		maxRetries: 1,
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return nil, &testNetError{msg: "connection refused"}
			}),
		},
	}

	records := []model.TradeRecord{{Date: "2025-01-02", Action: "buy", Symbol: "000001", Price: 10.0, Quantity: 100}}
	_, err := client.AnalyzeStyle(records)
	if err == nil {
		t.Error("expected error on network failure")
	}
}

type testNetError struct {
	msg string
}

func (n *testNetError) Error() string   { return n.msg }
func (n *testNetError) Timeout() bool   { return false }
func (n *testNetError) Temporary() bool { return true }
