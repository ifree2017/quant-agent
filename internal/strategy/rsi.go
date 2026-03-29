package strategy

import "quant-agent/internal/data"

// CalcRSI 计算RSI指标
func CalcRSI(closes []float64, period int) data.RSIData {
	if period <= 0 {
		period = 14
	}
	if len(closes) < period+1 {
		return data.RSIData{}
	}

	calcOne := func(p int) float64 {
		gains, losses := 0.0, 0.0
		for i := len(closes) - p; i < len(closes); i++ {
			diff := closes[i] - closes[i-1]
			if diff > 0 {
				gains += diff
			} else {
				losses += -diff
			}
		}
		avgGain := gains / float64(p)
		avgLoss := losses / float64(p)
		if avgLoss == 0 {
			return 100
		}
		rs := avgGain / avgLoss
		return 100 - (100 / (1 + rs))
	}

	return data.RSIData{
		RSI6:  calcOne(6),
		RSI14: calcOne(period),
	}
}
