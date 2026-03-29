package backtest

import "math"

// Account 模拟账户
type Account struct {
	InitialCash float64
	Cash       float64
	Locked     float64 // 持仓占用的资金
	PeakEquity float64
	History    []float64
}

// NewAccount 创建账户
func NewAccount(initialCash float64) *Account {
	return &Account{
		InitialCash: initialCash,
		Cash:       initialCash,
		Locked:     0,
		PeakEquity: initialCash,
	}
}

// Lock 锁定资金（买入持仓）
func (a *Account) Lock(amount float64) {
	a.Cash -= amount
	a.Locked += amount
}

// Unlock 解锁资金（卖出持仓）
func (a *Account) Unlock(amount float64) {
	a.Locked -= amount
	a.Cash += amount
}

// Add 增加资金（平仓盈利）
func (a *Account) Add(pnl float64) {
	a.Cash += pnl
	if a.Cash < 0 {
		a.Cash = 0
	}
}

// Equity 计算当前总权益
func (a *Account) Equity(portfolio *Portfolio) float64 {
	return a.Cash + portfolio.MarketValue(a.Cost())
}

// Cost 持仓成本
func (a *Account) Cost() float64 {
	return a.Locked
}

// UpdatePeak 更新权益峰值
func (a *Account) UpdatePeak(equity float64) {
	if equity > a.PeakEquity {
		a.PeakEquity = equity
	}
}

// MaxDrawdown 计算最大回撤
func (a *Account) MaxDrawdown() float64 {
	if len(a.History) == 0 {
		return 0
	}
	peak := a.InitialCash
	maxDD := 0.0
	for _, equity := range a.History {
		if equity > peak {
			peak = equity
		}
		dd := (peak - equity) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}
	return math.Max(0, maxDD)
}
