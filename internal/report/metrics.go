package report

import (
	"math"
	"quant-agent/internal/data"
	"quant-agent/internal/model"
	"time"
)

const (
	riskFreeRate = 0.03 // 无风险利率 3%
	tradingDays  = 252  // 年交易日
)

// CalcMetrics 计算绩效指标
func CalcMetrics(trades []model.Trade, initialCash, finalCash float64, bars []data.Bar) model.Metrics {
	if len(trades) == 0 {
		return model.Metrics{}
	}

	// 基础指标
	totalReturn := (finalCash - initialCash) / initialCash
	annualReturn := totalReturn * float64(tradingDays) / float64(len(bars))

	// 交易统计
	var wins, losses float64
	var totalPnL, totalWinPnL, totalLossPnL float64
	for _, t := range trades {
		totalPnL += t.PnL
		if t.PnL > 0 {
			wins++
			totalWinPnL += t.PnL
		} else {
			losses++
			totalLossPnL += t.PnL
		}
	}
	winRate := 0.0
	if len(trades) > 0 {
		winRate = wins / float64(len(trades))
	}
	avgWin := 0.0
	if wins > 0 {
		avgWin = totalWinPnL / wins
	}
	avgLoss := 0.0
	if losses > 0 {
		avgLoss = math.Abs(totalLossPnL / losses)
	}
	profitLossRatio := 0.0
	if avgLoss > 0 {
		profitLossRatio = avgWin / avgLoss
	}

	// 最大回撤
	maxDD := calcMaxDrawdown(trades, initialCash)

	// 夏普比率
	sharpe := calcSharpeRatio(trades, initialCash, annualReturn)

	// 卡玛比率
	calmar := 0.0
	if maxDD > 0 {
		calmar = annualReturn / maxDD
	}

	// 索提诺比率
	sortino := calcSortinoRatio(trades, initialCash, annualReturn)

	return model.Metrics{
		TotalReturn:     totalReturn,
		AnnualReturn:   annualReturn,
		SharpeRatio:    sharpe,
		MaxDrawdown:    maxDD,
		WinRate:        winRate,
		ProfitLossRatio: profitLossRatio,
		TotalTrades:    len(trades),
		CalmarRatio:    calmar,
		SortinoRatio:   sortino,
	}
}

// calcMaxDrawdown 计算最大回撤
func calcMaxDrawdown(trades []model.Trade, initialCash float64) float64 {
	if len(trades) == 0 {
		return 0
	}

	equity := initialCash
	peak := initialCash
	maxDD := 0.0

	for _, trade := range trades {
		equity += trade.PnL
		if equity > peak {
			peak = equity
		}
		dd := (peak - equity) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

// calcSharpeRatio 计算夏普比率
func calcSharpeRatio(trades []model.Trade, initialCash, annualReturn float64) float64 {
	if len(trades) < 2 {
		return 0
	}

	// 计算日收益序列
	var dailyReturns []float64
	equity := initialCash
	for _, trade := range trades {
		ret := trade.PnL / equity
		dailyReturns = append(dailyReturns, ret)
		equity += trade.PnL
	}

	if len(dailyReturns) < 2 {
		return 0
	}

	// 计算标准差
	mean := 0.0
	for _, r := range dailyReturns {
		mean += r
	}
	mean /= float64(len(dailyReturns))

	variance := 0.0
	for _, r := range dailyReturns {
		variance += math.Pow(r-mean, 2)
	}
	variance /= float64(len(dailyReturns))
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0
	}

	// 年化夏普比率
	annualizedReturn := mean * float64(tradingDays)
	annualizedStdDev := stdDev * math.Sqrt(float64(tradingDays))

	return (annualizedReturn - riskFreeRate) / annualizedStdDev
}

// calcSortinoRatio 计算索提诺比率
func calcSortinoRatio(trades []model.Trade, initialCash, annualReturn float64) float64 {
	if len(trades) < 2 {
		return 0
	}

	var downsideReturns []float64
	equity := initialCash
	for _, trade := range trades {
		ret := trade.PnL / equity
		if ret < 0 {
			downsideReturns = append(downsideReturns, ret)
		}
		equity += trade.PnL
	}

	if len(downsideReturns) == 0 {
		return annualReturn / riskFreeRate // 无下跌，返回高比率
	}

	downsideVariance := 0.0
	for _, r := range downsideReturns {
		downsideVariance += r * r
	}
	downsideStdDev := math.Sqrt(downsideVariance / float64(len(downsideReturns)))

	if downsideStdDev == 0 {
		return 0
	}

	annualizedReturn := annualReturn
	annualizedDownside := downsideStdDev * math.Sqrt(float64(tradingDays))

	return (annualizedReturn - riskFreeRate) / annualizedDownside
}

// Now 时间戳
func Now() time.Time {
	return time.Now()
}
