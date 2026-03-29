package backtest

import "quant-agent/internal/model"

// Portfolio 持仓组合
type Portfolio struct {
	Symbol      string
	Quantity    int
	entryDate  string
	entryPrice float64
}

// NewPortfolio 创建空组合
func NewPortfolio() *Portfolio {
	return &Portfolio{}
}

// HasPosition 是否有持仓
func (p *Portfolio) HasPosition() bool {
	return p.Quantity > 0
}

// Open 开仓
func (p *Portfolio) Open(date, symbol string, price float64, quantity int) {
	p.Symbol = symbol
	p.entryDate = date
	p.entryPrice = price
	p.Quantity = quantity
}

// Close 平仓
func (p *Portfolio) Close(exitDate string, exitPrice float64, reason string) model.Trade {
	if !p.HasPosition() {
		return model.Trade{}
	}
	pnl := (exitPrice - p.entryPrice) * float64(p.Quantity)
	pnlPct := (exitPrice - p.entryPrice) / p.entryPrice

	trade := model.Trade{
		EntryDate:  p.entryDate,
		ExitDate:   exitDate,
		Symbol:     p.Symbol,
		EntryPrice: p.entryPrice,
		ExitPrice:  exitPrice,
		Quantity:   p.Quantity,
		PnL:        pnl,
		PnLPct:     pnlPct,
		ExitReason: reason,
	}
	p.Quantity = 0
	return trade
}

// MarketValue 计算市值
func (p *Portfolio) MarketValue(currentPrice float64) float64 {
	if !p.HasPosition() {
		return 0
	}
	return currentPrice * float64(p.Quantity)
}

// EntryPriceValue 持仓价格
func (p *Portfolio) EntryPriceValue() float64 {
	return p.entryPrice
}

// Cost 持仓成本
func (p *Portfolio) Cost() float64 {
	if !p.HasPosition() {
		return 0
	}
	return p.entryPrice * float64(p.Quantity)
}
