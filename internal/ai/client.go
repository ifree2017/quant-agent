package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"quant-agent/internal/model"
)

// Client LLM API客户端
type Client struct {
	baseURL    string
	token     string
	model     string
	client    *http.Client
	maxRetries int
}

// NewClient 创建客户端
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:     token,
		model:     "gpt-4o-mini",
		client:    &http.Client{Timeout: 60 * time.Second},
		maxRetries: 3,
	}
}

// AnalyzeStyle 分析用户风格
func (c *Client) AnalyzeStyle(records []model.TradeRecord) (model.StyleProfile, error) {
	prompt := buildStylePrompt(records)
	return callLLM[model.StyleProfile](prompt, c)
}

// GenerateStrategy 生成策略
func (c *Client) GenerateStrategy(profile model.StyleProfile) (model.StrategyRules, error) {
	prompt := buildStrategyPrompt(profile)
	return callLLM[model.StrategyRules](prompt, c)
}

// callLLM 调用LLM并解析响应
func callLLM[T any](prompt string, client *Client) (T, error) {
	var zero T

	payload := map[string]any{
		"model": client.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return zero, fmt.Errorf("marshal: %w", err)
	}

	var respBytes []byte
	var lastErr error
	for i := 0; i < client.maxRetries; i++ {
		req, _ := http.NewRequest("POST", client.baseURL+"/chat/completions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+client.token)

		resp, err := client.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 2 * time.Second) // 指数退避
			continue
		}
		defer resp.Body.Close()

		respBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("llm status: %d, body: %s", resp.StatusCode, string(respBytes))
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}

		lastErr = nil
		break
	}
	if lastErr != nil {
		return zero, fmt.Errorf("llm call failed after retries: %w", lastErr)
	}

	// 解析响应
	var llmResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBytes, &llmResp); err != nil {
		return zero, fmt.Errorf("parse llm response: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return zero, fmt.Errorf("llm no choices")
	}

	content := llmResp.Choices[0].Message.Content

	// 尝试提取JSON
	content = extractJSON(content)

	var result T
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return zero, fmt.Errorf("parse result: %w, content: %s", err, content)
	}

	return result, nil
}

// buildStylePrompt 构建风格分析prompt
func buildStylePrompt(records []model.TradeRecord) string {
	recordsJSON, _ := json.Marshal(records)
	return fmt.Sprintf(`你是一个量化投资风格分析师。根据用户的历史交易记录，分析出风格画像。

交易记录：
%s

输出严格JSON格式（无其他内容）：
{
  "style": "保守|稳健|平衡|积极|激进",
  "riskScore": 0-100,
  "tradeFrequency": "高频|中频|低频",
  "avgHoldDays": 平均持仓天数,
  "maxDrawdownTolerance": 可接受最大回撤百分比
}`, recordsJSON)
}

// buildStrategyPrompt 构建策略生成prompt
func buildStrategyPrompt(profile model.StyleProfile) string {
	profileJSON, _ := json.Marshal(profile)
	return fmt.Sprintf(`你是一个量化策略专家。根据用户的风格画像，生成量化交易策略规则。

风格画像：
%s

输出严格JSON格式（无其他内容）：
{
  "indicators": ["MA","RSI","MACD","Bollinger"],
  "params": {
    "MA_period": 20,
    "RSI_period": 14,
    "RSI_overbought": 70,
    "RSI_oversold": 30,
    "MACD_fast": 12,
    "MACD_slow": 26,
    "MACD_signal": 9
  },
  "entry": {"type": "cross|indicator_value", "condition": "MA_cross_RSI|RSI_oversold|MACD_cross_signal"},
  "exit": {"type": "stop_loss", "value": 0.05},
  "position": {"type": "fixed", "value": 0.2},
  "market": "A-share"
}`, profileJSON)
}

// extractJSON 从文本中提取JSON
func extractJSON(s string) string {
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '{' {
			start = i
			break
		}
	}
	if start > 0 {
		return s[start:]
	}
	return s
}
