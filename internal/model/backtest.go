package model

import "time"

// BacktestConfig 回测配置
type BacktestConfig struct {
	StrategyID   string  `json:"strategyId"`
	Symbol      string  `json:"symbol"`
	Days        int     `json:"days"`         // 回测天数
	InitialCash float64 `json:"initialCash"`   // 初始资金
}

// BacktestResult 回测结果
type BacktestResult struct {
	ID            string      `json:"id"`
	StrategyID    string      `json:"strategyId"`
	Symbol       string      `json:"symbol"`
	Days          int         `json:"days"`
	InitialCash  float64     `json:"initialCash"`
	FinalCash    float64     `json:"finalCash"`
	Metrics       Metrics     `json:"metrics"`
	Trades       []Trade     `json:"trades"`
	EquityCurve  []EquityPoint `json:"equityCurve"`
	CreatedAt    time.Time   `json:"createdAt"`
}

// Trade 成交记录
type Trade struct {
	ID         int     `json:"id"`
	EntryDate  string  `json:"entryDate"`
	ExitDate   string  `json:"exitDate"`
	Symbol     string  `json:"symbol"`
	EntryPrice float64 `json:"entryPrice"`
	ExitPrice  float64 `json:"exitPrice"`
	Quantity   int     `json:"quantity"`
	PnL        float64 `json:"pnl"`        // 盈亏金额
	PnLPct     float64 `json:"pnlPct"`    // 盈亏比例
	ExitReason string  `json:"exitReason"` // stop_loss/take_profit/signal
}

// EquityPoint 权益曲线点
type EquityPoint struct {
	Date     string  `json:"date"`
	Equity   float64 `json:"equity"`
	Drawdown float64 `json:"drawdown"` // 回撤比例
}

// Metrics 绩效指标
type Metrics struct {
	TotalReturn    float64 `json:"totalReturn"`    // 总收益率
	AnnualReturn  float64 `json:"annualReturn"`   // 年化收益率
	SharpeRatio   float64 `json:"sharpeRatio"`   // 夏普比率
	MaxDrawdown   float64 `json:"maxDrawdown"`   // 最大回撤 %
	WinRate       float64 `json:"winRate"`       // 胜率
	ProfitLossRatio float64 `json:"profitLossRatio"` // 盈亏比
	TotalTrades   int     `json:"totalTrades"`   // 总交易次数
	CalmarRatio   float64 `json:"calmarRatio"`   // 卡玛比率
	SortinoRatio  float64 `json:"sortinoRatio"`  // 索提诺比率
}

// BacktestRunRequest 执行回测请求
type BacktestRunRequest struct {
	StrategyID  string `json:"strategyId" binding:"required"`
	Symbol     string `json:"symbol" binding:"required"`
	Days       int    `json:"days"`        // 默认 60
	InitialCash float64 `json:"initialCash"` // 默认 50000
}
