package handler

import (
	"net/http"

	"quant-agent/internal/data"
	"quant-agent/pkg/signal"

	"github.com/gin-gonic/gin"
)

// GetDistribution 获取出货识别信号
// GET /api/v1/distribution/:code
func GetDistribution() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "stock code is required",
			})
			return
		}

		// 获取分时/K线数据
		features, err := fetchDistributionFeatures(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		// 计算信号
		distSignal := signal.NewDistributionSignal(code, features)

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"code":       distSignal.Code,
				"signal":     distSignal.Signal,
				"score":      distSignal.Score,
				"features": gin.H{
					"spike_fall":        distSignal.Features.SpikeFall,
					"second_wave":       distSignal.Features.SecondWave,
					"volume_decay":      distSignal.Features.VolumeDecay,
					"sell_pressure":      distSignal.Features.SellPressure,
					"active_sell_ratio": distSignal.Features.ActiveSellRatio,
				},
				"action":    distSignal.Action,
				"timestamp": distSignal.Timestamp.Format("2006-01-02T15:04:05Z"),
			},
		})
	}
}

// fetchDistributionFeatures 获取出货特征
// 优先使用分时数据，如果不可用则使用K线数据模拟
func fetchDistributionFeatures(code string) (signal.DistributionFeatures, error) {
	// 尝试获取分时数据
	bars, err := fetchTodayBars(code)
	if err != nil || len(bars) == 0 {
		// 回退到K线数据模拟
		return simulateDistributionFeatures(code)
	}

	return extractDistributionFeatures(bars)
}

// extractDistributionFeatures 从K线数据提取出货特征
func extractDistributionFeatures(bars []data.Bar) (signal.DistributionFeatures, error) {
	if len(bars) < 2 {
		return signal.DistributionFeatures{}, nil
	}

	features := signal.DistributionFeatures{}

	// 取最新bar和前一个bar
	lastBar := bars[len(bars)-1]
	prevBar := bars[len(bars)-2]

	// 计算当日最高/最低
	var dayHigh, dayLow float64
	var morningVol, afternoonVol int64
	halfPoint := len(bars) / 2

	for i, bar := range bars {
		if bar.High > dayHigh || dayHigh == 0 {
			dayHigh = bar.High
		}
		if bar.Low < dayLow || dayLow == 0 {
			dayLow = bar.Low
		}
		// 简单区分上下半场（按时间顺序）
		if i < halfPoint {
			morningVol += bar.Volume
		} else {
			afternoonVol += bar.Volume
		}
	}

	// spike_fall: 冲高回落（>3%拉高后回落）
	openPrice := bars[0].Open
	if dayHigh > 0 && openPrice > 0 {
		risePct := (dayHigh - openPrice) / openPrice * 100
		fallPct := (dayHigh - lastBar.Close) / dayHigh * 100
		features.SpikeFall = risePct > 3 && fallPct > 1
	}

	// second_wave: 有无二波攻击（通过价格反弹判断）
	// 如果价格在下午创出新高，认为有二波
	midPrice := bars[halfPoint].Close
	features.SecondWave = lastBar.Close > midPrice && lastBar.Close > prevBar.Close

	// volume_decay: 量能衰减（下半场<上半场70%）
	if morningVol > 0 {
		features.VolumeDecay = float64(afternoonVol) < float64(morningVol)*0.7
	}

	// sell_pressure: 卖盘压制（通过价格下跌幅度判断）
	// 假设：收盘价在低位运行表示卖盘压制
	dayRange := dayHigh - dayLow
	if dayRange > 0 {
		closePosition := (dayHigh - lastBar.Close) / dayRange
		features.SellPressure = closePosition > 0.7 // 收盘价在当日 range 的70%以上位置（即从高位回落）
	}

	// active_sell_ratio: 主动卖出比（简化估算）
	// 通过涨跌判断：下跌日主动卖出多
	if lastBar.Close < prevBar.Close {
		features.ActiveSellRatio = 0.7
	} else if lastBar.Close > prevBar.Close {
		features.ActiveSellRatio = 1.2
	} else {
		features.ActiveSellRatio = 1.0
	}

	return features, nil
}

// simulateDistributionFeatures 使用历史K线模拟出货特征
func simulateDistributionFeatures(code string) (signal.DistributionFeatures, error) {
	// 获取近期K线数据
	bars, err := fetchRecentBars(code, 20)
	if err != nil || len(bars) < 5 {
		return signal.DistributionFeatures{}, err
	}

	return extractDistributionFeatures(bars)
}

// fetchTodayBars 获取当日分时/K线数据
func fetchTodayBars(code string) ([]data.Bar, error) {
	// 尝试从东财API获取实时数据
	// 这里简化处理，实际应调用 social/eastmoney 获取分时数据
	return nil, nil
}

// fetchRecentBars 获取近期K线数据
func fetchRecentBars(code string, days int) ([]data.Bar, error) {
	// 这里应调用数据加载模块获取历史K线
	// 暂时返回空，外部会使用默认特征
	return nil, nil
}

// GetBuyPoint 获取买点识别信号
// GET /api/v1/buypoint/:code
func GetBuyPoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "stock code is required",
			})
			return
		}

		// 获取K线数据
		bars, err := fetchBuyPointBars(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		// 提取买点模式
		patterns := extractBuyPointPatterns(bars)

		// 防追涨检查
		antiChase := checkAntiChase(bars)

		// 计算信号
		buySignal := signal.NewBuyPointSignal(code, patterns, antiChase)

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"code":       buySignal.Code,
				"signal":     buySignal.Signal,
				"score":      buySignal.Score,
				"patterns": gin.H{
					"pullback":    buySignal.Patterns.Pullback,
					"breakout":    buySignal.Patterns.Breakout,
					"second_wave": buySignal.Patterns.SecondWave,
					"flow":        buySignal.Patterns.Flow,
					"pressure":    buySignal.Patterns.Pressure,
				},
				"anti_chase": buySignal.AntiChase,
				"action":     buySignal.Action,
				"timestamp":  buySignal.Timestamp.Format("2006-01-02T15:04:05Z"),
			},
		})
	}
}

// fetchBuyPointBars 获取买点分析所需的K线数据
func fetchBuyPointBars(code string) ([]data.Bar, error) {
	// 获取近期K线数据（至少20天用于计算均量）
	return fetchRecentBars(code, 20)
}

// extractBuyPointPatterns 从K线数据提取买点模式
func extractBuyPointPatterns(bars []data.Bar) signal.BuyPointPatterns {
	patterns := signal.BuyPointPatterns{}

	if len(bars) < 5 {
		return patterns
	}

	lastBar := bars[len(bars)-1]
	prevBar := bars[len(bars)-2]

	// 计算均线（简化：使用5日均线作为参考）
	ma5 := calculateMA(bars, 5)
	ma10 := calculateMA(bars, 10)

	// 计算均量
	avgVol := calculateAvgVolume(bars, 5)

	// 计算前高（前20日最高）
	prevHigh := calculatePrevHigh(bars, 20)

	// pullback: 回踩幅度2-5%，不破均线
	patterns.Pullback = signal.CheckPullback(lastBar.Close, ma5) || signal.CheckPullback(lastBar.Close, ma10)

	// breakout: 突破前高+放量
	patterns.Breakout = signal.CheckBreakout(lastBar.High, prevHigh, float64(lastBar.Volume), avgVol)

	// second_wave: 接近前高+量能再放大
	patterns.SecondWave = signal.CheckSecondWave(lastBar.Close, prevHigh, float64(lastBar.Volume), float64(prevBar.Volume))

	// flow: 主动买入比（简化：以上涨日为买入占优）
	patterns.Flow = lastBar.Close > prevBar.Close

	// pressure: 无卖盘压制（简化：收盘在高位表示抛压小）
	dayRange := lastBar.High - lastBar.Low
	if dayRange > 0 {
		closePos := (lastBar.Close - lastBar.Low) / dayRange
		patterns.Pressure = closePos > 0.5
	}

	return patterns
}

// checkAntiChase 防追涨检查
func checkAntiChase(bars []data.Bar) bool {
	if len(bars) < 2 {
		return false
	}

	// 转换为 interface{} 切片
	interfaceBars := make([]interface{}, len(bars))
	for i, bar := range bars {
		interfaceBars[i] = bar
	}

	return signal.AntiChase(interfaceBars)
}

// calculateMA 计算移动平均
func calculateMA(bars []data.Bar, period int) float64 {
	if len(bars) < period {
		return 0
	}
	sum := 0.0
	for i := len(bars) - period; i < len(bars); i++ {
		sum += bars[i].Close
	}
	return sum / float64(period)
}

// calculateAvgVolume 计算平均成交量
func calculateAvgVolume(bars []data.Bar, period int) float64 {
	if len(bars) < period {
		return 0
	}
	sum := int64(0)
	for i := len(bars) - period; i < len(bars); i++ {
		sum += bars[i].Volume
	}
	return float64(sum) / float64(period)
}

// calculatePrevHigh 计算前高（排除当前bar）
func calculatePrevHigh(bars []data.Bar, period int) float64 {
	if len(bars) < period+1 {
		return 0
	}
	// 取前 period 根 bar 的最高价
	var high float64
	for i := len(bars) - period - 1; i < len(bars)-1; i++ {
		if bars[i].High > high {
			high = bars[i].High
		}
	}
	return high
}
