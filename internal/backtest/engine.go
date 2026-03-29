package backtest

import (
	"fmt"
	"quant-agent/internal/data"
	"quant-agent/internal/model"
	"quant-agent/internal/report"
	"quant-agent/internal/strategy"
	"time"

	"github.com/google/uuid"
)

// Engine 回测引擎
type Engine struct {
	cfg  model.BacktestConfig
	data *data.Loader
}

// NewEngine 创建回测引擎
func NewEngine(cfg model.BacktestConfig, dataDir string) *Engine {
	return &Engine{
		cfg:  cfg,
		data: data.NewLoader(dataDir),
	}
}

// Run 执行回测
func (e *Engine) Run(rules model.StrategyRules, bars []data.Bar) (*model.BacktestResult, error) {
	if len(bars) < 2 {
		return nil, fmt.Errorf("bars too few: %d", len(bars))
	}

	account := NewAccount(e.cfg.InitialCash)
	executor := strategy.NewExecutor(rules)
	portfolio := NewPortfolio()

	var trades []model.Trade
	var equityCurve []model.EquityPoint

	for i, bar := range bars {
		executor.Update(bar)

		if i == 0 {
			equityCurve = append(equityCurve, model.EquityPoint{
				Date:     bar.Date.Format("2006-01-02"),
				Equity:   account.Cash,
				Drawdown: 0,
			})
			continue
		}

		signal := executor.Signal()
		price := bar.Close
		positionSize := e.calcPosition(rules, account.Cash, price)

		if signal == "buy" && !portfolio.HasPosition() && positionSize > 0 {
			cost := price * float64(positionSize)
			if cost <= account.Cash {
				entryDate := executor.Store().Bars[len(executor.Store().Bars)-1].Date.Format("2006-01-02")
				portfolio.Open(entryDate, e.cfg.Symbol, price, positionSize)
				account.Lock(cost)
			}
		}

		if signal == "sell" && portfolio.HasPosition() {
			exitReason := e.calcExitReason(rules, bar, portfolio)
			trade := portfolio.Close(bar.Date.Format("2006-01-02"), price, exitReason)
			if trade.PnL != 0 || true {
				account.Unlock(portfolio.Cost())
				account.Add(trade.PnL)
				trades = append(trades, trade)
			}
		}

		// 止损检查
		if portfolio.HasPosition() {
			exitReason := e.checkStopLoss(rules, bar, portfolio)
			if exitReason != "" {
				trade := portfolio.Close(bar.Date.Format("2006-01-02"), price, exitReason)
				account.Unlock(portfolio.Cost())
				account.Add(trade.PnL)
				trades = append(trades, trade)
			}
		}

		equity := account.Equity(portfolio)
		peak := account.PeakEquity
		drawdown := 0.0
		if peak > 0 {
			drawdown = (peak - equity) / peak
		}
		equityCurve = append(equityCurve, model.EquityPoint{
			Date:     bar.Date.Format("2006-01-02"),
			Equity:   equity,
			Drawdown: drawdown,
		})
		account.UpdatePeak(equity)
	}

	// 平仓
	if portfolio.HasPosition() {
		lastBar := bars[len(bars)-1]
		trade := portfolio.Close(lastBar.Date.Format("2006-01-02"), lastBar.Close, "end_of_backtest")
		account.Add(trade.PnL)
		trades = append(trades, trade)
	}

	metrics := report.CalcMetrics(trades, account.InitialCash, account.Equity(portfolio), bars)

	result := &model.BacktestResult{
		ID:           uuid.New().String(),
		StrategyID:   e.cfg.StrategyID,
		Symbol:       e.cfg.Symbol,
		Days:         len(bars),
		InitialCash:  account.InitialCash,
		FinalCash:   account.Equity(portfolio),
		Metrics:     metrics,
		Trades:      trades,
		EquityCurve: equityCurve,
		CreatedAt:   time.Now(),
	}

	return result, nil
}

// calcPosition 计算仓位数量
func (e *Engine) calcPosition(rules model.StrategyRules, cash float64, price float64) int {
	if rules.Position.Type == "fixed" {
		value := cash * rules.Position.Value
		return int(value / price)
	}
	return int(cash * 0.2 / price)
}

// calcExitReason 计算退出原因
func (e *Engine) calcExitReason(rules model.StrategyRules, bar data.Bar, portfolio *Portfolio) string {
	if rules.Exit.Type == "stop_loss" && portfolio.HasPosition() {
		entry := portfolio.EntryPriceValue()
		lossPct := (entry - bar.Close) / entry
		if lossPct > rules.Exit.Value {
			return "stop_loss"
		}
	}
	return "signal"
}

// checkStopLoss 止损检查
func (e *Engine) checkStopLoss(rules model.StrategyRules, bar data.Bar, portfolio *Portfolio) string {
	if !portfolio.HasPosition() {
		return ""
	}
	entry := portfolio.EntryPriceValue()
	lossPct := (entry - bar.Close) / entry
	if lossPct > rules.Exit.Value {
		return "stop_loss"
	}
	return ""
}
