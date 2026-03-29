package model

import "time"

// Strategy 量化策略
type Strategy struct {
	ID        string    `json:"id"`
	UserID   string    `json:"userId"`
	Name     string    `json:"name"`
	Style    string    `json:"style"`    // 对应风格
	Rules    StrategyRules `json:"rules"`
	Version  int       `json:"version"`
	Tags     []string `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// StrategyRules 策略规则
type StrategyRules struct {
	Indicators []string       `json:"indicators"` // 指标列表: MA/RSI/MACD/Bollinger
	Params    map[string]float64 `json:"params"`     // 参数
	Entry     EntryRule       `json:"entry"`      // 入场条件
	Exit      ExitRule        `json:"exit"`       // 出场条件
	Position  PositionRule    `json:"position"`   // 仓位管理
	Market    string          `json:"market"`     // 市场: A-share/Futures/Crypto
}

// EntryRule 入场规则
type EntryRule struct {
	Type      string `json:"type"`       // cross/indicator_value/time
	Condition string `json:"condition"`   // 具体条件描述
}

// ExitRule 出场规则
type ExitRule struct {
	Type  string  `json:"type"`  // stop_loss/take_profit/time/stop_loss_pct
	Value float64 `json:"value"`
}

// PositionRule 仓位规则
type PositionRule struct {
	Type  string  `json:"type"`  // fixed/risk_based
	Value float64 `json:"value"` // 固定仓位比例 或 风险比例
}

// StrategyGenerateRequest 策略生成请求
type StrategyGenerateRequest struct {
	UserID       string      `json:"userId" binding:"required"`
	StyleProfile StyleProfile `json:"styleProfile" binding:"required"`
	Market      string      `json:"market"`    // 默认 A-share
	Symbol      string      `json:"symbol"`    // 可选
}

// StrategyReport 策略报告（生成策略 + 回测结果）
type StrategyReport struct {
	Strategy     Strategy     `json:"strategy"`
	Backtest    *BacktestResult `json:"backtest,omitempty"`
}
