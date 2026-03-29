package handler

import (
	"net/http"

	"quant-agent/pkg/signal"

	"github.com/gin-gonic/gin"
)

var (
	signalEngine *signal.Engine
)

// getSignalEngine 获取或创建信号引擎
func getSignalEngine() *signal.Engine {
	if signalEngine == nil {
		signalEngine = signal.NewEngine()
	}
	return signalEngine
}

// GetSignal 获取个股综合交易信号
// GET /api/v1/signal/:code
// 返回 Body: { stock_code, signal_type, confidence, score, message, created_at }
func GetSignal() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "stock code is required",
			})
			return
		}

		engine := getSignalEngine()
		signals, err := engine.GenerateSignals(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		if len(signals) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"data":    nil,
				"message": "no signal generated",
			})
			return
		}

		// 返回第一个信号
		s := signals[0]
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"stock_code":  s.StockCode,
				"signal_type": s.SignalType,
				"confidence":  s.Confidence,
				"score":       s.Score,
				"source":      s.Source,
				"message":     s.Message,
				"created_at":   s.CreatedAt.Format("2006-01-02 15:04:05"),
			},
		})
	}
}

// GetSectorSignals 获取板块热点轮动榜单
// GET /api/v1/signal/sector
// 返回 Body: [{ sector_name, heat_score, direction, keywords, source, created_at }]
func GetSectorSignals() gin.HandlerFunc {
	return func(c *gin.Context) {
		engine := getSignalEngine()
		signals, err := engine.DetectSectorHot()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		// 转换输出格式
		data := make([]gin.H, 0, len(signals))
		for _, s := range signals {
			data = append(data, gin.H{
				"sector_name": s.SectorName,
				"heat_score":  s.HeatScore,
				"direction":   s.Direction,
				"stock_codes": s.StockCodes,
				"change_pct":  s.ChangePct,
				"keywords":    s.Keywords,
				"source":      s.Source,
				"created_at":  s.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": data,
		})
	}
}

// GetAlerts 获取市场舆情预警列表（全局）
// GET /api/v1/signal/alerts
// 返回 Body: [{ stock_code, alert_level, title, description, keywords, news_count, duration, source, created_at }]
// 注意：这个API返回全局预警，需要指定stock_code参数过滤
func GetAlerts() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("stock_code")
		if code == "" {
			// 如果没有指定股票代码，返回所有市场预警（需要从各板块综合判断）
			// 这里简化处理，返回全局预警为空数组，用户需要指定股票
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "stock_code parameter is required",
			})
			return
		}

		engine := getSignalEngine()
		alerts, err := engine.DetectAlerts(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		// 转换输出格式
		data := make([]gin.H, 0, len(alerts))
		for _, a := range alerts {
			data = append(data, gin.H{
				"stock_code":   a.StockCode,
				"alert_level":  a.AlertLevel,
				"title":        a.Title,
				"description":  a.Description,
				"keywords":     a.Keywords,
				"news_count":   a.NewsCount,
				"duration":     a.Duration,
				"source":       a.Source,
				"created_at":   a.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": data,
		})
	}
}
