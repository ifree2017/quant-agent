package strategy

import (
	"quant-agent/internal/data"
	"quant-agent/internal/model"
)

// Executor 策略执行器
type Executor struct {
	rules  model.StrategyRules
	store  *data.IndicatorStore
	bar    data.Bar
	signal string // buy/sell/hold
}

// NewExecutor 创建执行器
func NewExecutor(rules model.StrategyRules) *Executor {
	return &Executor{
		rules:  rules,
		store:  &data.IndicatorStore{},
	}
}

// Update 更新最新bar并计算指标
func (e *Executor) Update(bar data.Bar) {
	e.bar = bar
	e.store.Bars = append(e.store.Bars, bar)
	e.store.ClosePrices = append(e.store.ClosePrices, bar.Close)

	closes := e.store.ClosePrices

	if contains(e.rules.Indicators, "MA") {
		e.store.MA = CalcMA(closes, e.rules.Params)
	}
	if contains(e.rules.Indicators, "RSI") {
		e.store.RSI = CalcRSI(closes, int(e.rules.Params["RSI_period"]))
	}
	if contains(e.rules.Indicators, "MACD") {
		e.store.MACD = CalcMACD(closes,
			int(e.rules.Params["MACD_fast"]),
			int(e.rules.Params["MACD_slow"]),
			int(e.rules.Params["MACD_signal"]))
	}
	if contains(e.rules.Indicators, "Bollinger") {
		e.store.Bollinger = CalcBollinger(closes, 20)
	}
}

// Signal 生成交易信号
func (e *Executor) Signal() string {
	rules := e.rules
	closes := e.store.ClosePrices
	if len(closes) < 2 {
		return "hold"
	}

	switch rules.Entry.Condition {
	case "MA_cross_RSI":
		maVal := e.store.MA.MA20
		rsiVal := e.store.RSI.RSI14
		if maVal > 0 && rsiVal > 0 {
			if closes[len(closes)-1] > maVal && rsiVal < 70 {
				return "buy"
			}
			if rsiVal > 70 || closes[len(closes)-1] < maVal {
				return "sell"
			}
		}
	case "RSI_oversold":
		if e.store.RSI.RSI14 > 0 && e.store.RSI.RSI14 < float64(int(e.rules.Params["RSI_oversold"])) {
			return "buy"
		}
		if e.store.RSI.RSI14 > float64(int(e.rules.Params["RSI_overbought"])) {
			return "sell"
		}
	case "MACD_cross_signal":
		if e.store.MACD.Histogram > 0 {
			return "buy"
		}
		if e.store.MACD.Histogram < 0 {
			return "sell"
		}
	default:
		if e.store.MA.MA20 > 0 && closes[len(closes)-1] > e.store.MA.MA20 {
			return "buy"
		}
		if e.store.MA.MA20 > 0 && closes[len(closes)-1] < e.store.MA.MA20 {
			return "sell"
		}
	}

	return "hold"
}

// Store 返回指标存储（用于外部访问）
func (e *Executor) Store() *data.IndicatorStore {
	return e.store
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
