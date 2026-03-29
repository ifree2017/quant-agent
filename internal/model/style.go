package model

import "time"

// StyleProfile 用户操作风格画像
type StyleProfile struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"userId"`
	Style                 string    `json:"style"` // 保守/稳健/平衡/积极/激进
	RiskScore             float64   `json:"riskScore"`              // 0-100
	TradeFrequency        string    `json:"tradeFrequency"`         // 高频/中频/低频
	AvgHoldDays           float64   `json:"avgHoldDays"`            // 平均持仓天数
	MaxDrawdownTolerance  float64   `json:"maxDrawdownTolerance"`  // 可接受最大回撤 %
	CreatedAt             time.Time `json:"createdAt"`
}

// TradeRecord 用户历史交易记录
type TradeRecord struct {
	Date     string  `json:"date"`      // 格式: 2025-01-02
	Action   string  `json:"action"`    // buy/sell
	Symbol   string  `json:"symbol"`     // 股票代码
	Price    float64 `json:"price"`      // 成交价格
	Quantity int     `json:"quantity"`   // 成交数量
}

// StyleAnalyzeRequest 风格分析请求
type StyleAnalyzeRequest struct {
	UserID  string        `json:"userId" binding:"required"`
	Records []TradeRecord `json:"records" binding:"required,min=1"`
}
