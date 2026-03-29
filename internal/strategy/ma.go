package strategy

import "quant-agent/internal/data"

// CalcMA 计算移动平均线
func CalcMA(closes []float64, params map[string]float64) data.MAData {
	n5 := 5
	n10 := 10
	n20 := 20
	n60 := 60

	if period, ok := params["MA_period"]; ok {
		n20 = int(period)
	}

	return data.MAData{
		MA5:  sma(closes, n5),
		MA10: sma(closes, n10),
		MA20: sma(closes, n20),
		MA60: sma(closes, n60),
	}
}

func sma(closes []float64, n int) float64 {
	if len(closes) < n {
		return 0
	}
	sum := 0.0
	for i := len(closes) - n; i < len(closes); i++ {
		sum += closes[i]
	}
	return sum / float64(n)
}
