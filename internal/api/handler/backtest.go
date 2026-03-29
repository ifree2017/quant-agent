package handler

import (
	"net/http"
	"quant-agent/internal/backtest"
	"quant-agent/internal/data"
	"quant-agent/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BacktestRun 执行回测
func BacktestRun(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.BacktestRunRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Days == 0 {
			req.Days = 60
		}
		if req.InitialCash == 0 {
			req.InitialCash = 50000
		}

		cfg := model.BacktestConfig{
			StrategyID:  req.StrategyID,
			Symbol:     req.Symbol,
			Days:       req.Days,
			InitialCash: req.InitialCash,
		}

		// TODO: 从存储层获取strategy的rules
		defaultRules := model.StrategyRules{
			Indicators: []string{"MA", "RSI"},
			Params: map[string]float64{
				"MA_period":      20,
				"RSI_period":     14,
				"RSI_overbought": 70,
				"RSI_oversold":   30,
			},
		}

		go func() {
			loader := data.NewLoader(dataDir)
			bars, _ := loader.LoadBars(cfg.Symbol, cfg.Days)
			engine := backtest.NewEngine(cfg, dataDir)
			_, _ = engine.Run(defaultRules, bars)
		}()

		c.JSON(http.StatusOK, gin.H{
			"jobId": uuid.New().String(),
			"status": "queued",
		})
	}
}

// BacktestGet 回测结果
func BacktestGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"result": nil})
	}
}
