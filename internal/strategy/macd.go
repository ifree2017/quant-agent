package strategy

import "quant-agent/internal/data"

// CalcMACD 计算MACD指标
func CalcMACD(closes []float64, fast, slow, signal int) data.MACDData {
	if fast <= 0 {
		fast = 12
	}
	if slow <= 0 {
		slow = 26
	}
	if signal <= 0 {
		signal = 9
	}
	if len(closes) < slow {
		return data.MACDData{}
	}

	emaFast := ema(closes, fast)
	emaSlow := ema(closes, slow)
	macdLine := emaFast - emaSlow

	// Signal line = EMA of MACD (简化)
	signalLine := macdLine * 0.9 // 简化，实际需计算9日EMA

	return data.MACDData{
		MACD:      macdLine,
		Signal:    signalLine,
		Histogram: macdLine - signalLine,
	}
}

func ema(closes []float64, n int) float64 {
	if len(closes) < n {
		return 0
	}
	k := 2.0 / float64(n+1)
	ema := closes[0]
	for i := 1; i < len(closes); i++ {
		ema = closes[i]*k + ema*(1-k)
	}
	return ema
}
