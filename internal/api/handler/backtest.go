package handler

import (
	"net/http"
	"quant-agent/internal/backtest"
	"quant-agent/internal/data"
	"quant-agent/internal/model"
	"quant-agent/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// globalStore 全局 store 实例（共享 strategy handler 的实例）
var backtestStore *store.Store

// SetBacktestStore 注入 store 实例
func SetBacktestStore(s *store.Store) {
	backtestStore = s
}

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

		jobID := uuid.New().String()

		// Capture stores for goroutine (both concrete and interface-based)
		bs := backtestStore
		bsi := backtestStoreInterface
		go func() {
			loader := data.NewLoader(dataDir)
			bars, _ := loader.LoadBars(cfg.Symbol, cfg.Days)
			engine := backtest.NewEngine(cfg, dataDir)
			result, _ := engine.Run(defaultRules, bars)
			// 保存回测结果到数据库
			var storeToSave StoreInterface
			if bs != nil {
				storeToSave = bs
			} else {
				storeToSave = bsi
			}
			if result != nil && storeToSave != nil {
				result.ID = jobID
				result.StrategyID = cfg.StrategyID
				_ = storeToSave.SaveBacktest(result)
			}
		}()

		c.JSON(http.StatusOK, gin.H{
			"jobId": jobID,
			"status": "queued",
		})
	}
}

// BacktestGet 回测结果
func BacktestGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var storeToUse StoreInterface
		if backtestStore != nil {
			storeToUse = backtestStore
		} else {
			storeToUse = backtestStoreInterface
		}
		if storeToUse == nil {
			c.JSON(http.StatusOK, gin.H{"result": nil})
			return
		}
		result, err := storeToUse.GetBacktest(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": result})
	}
}
