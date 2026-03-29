package strategy

import (
	"math"
	"quant-agent/internal/data"
)

// CalcBollinger 计算布林带
func CalcBollinger(closes []float64, period int) data.BollingerData {
	if len(closes) < period {
		return data.BollingerData{}
	}
	middle := sma(closes, period)
	std := stddev(closes[len(closes)-period:], period)
	upper := middle + 2*std
	lower := middle - 2*std
	return data.BollingerData{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
	}
}

func stddev(vals []float64, n int) float64 {
	if len(vals) < n {
		return 0
	}
	mean := sma(vals, n)
	var sum float64
	for _, v := range vals[len(vals)-n:] {
		sum += math.Pow(v-mean, 2)
	}
	return math.Sqrt(sum / float64(n))
}
