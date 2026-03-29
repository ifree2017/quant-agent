package handler

import (
	"net/http"
	"quant-agent/internal/ai"
	"quant-agent/internal/backtest"
	"quant-agent/internal/data"
	"quant-agent/internal/model"
	"quant-agent/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// globalStore 全局 store 实例
var globalStore *store.Store

// SetStore 注入 store 实例
func SetStore(s *store.Store) {
	globalStore = s
}

// StrategyGenerate 策略生成
func StrategyGenerate(aiClient *ai.Client, dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.StrategyGenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		rules, err := aiClient.GenerateStrategy(req.StyleProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		strategy := &model.Strategy{
			ID:    uuid.New().String(),
			UserID: req.UserID,
			Name:   req.StyleProfile.Style + "型策略",
			Style: req.StyleProfile.Style,
			Rules: rules,
		}

		// 自动回测
		cfg := model.BacktestConfig{
			StrategyID: strategy.ID,
			Symbol:    req.Symbol,
			Days:      60,
		}
		var report *model.BacktestResult
		func() {
			loader := data.NewLoader(dataDir)
			bars, _ := loader.LoadBars(req.Symbol, 60)
			engine := backtest.NewEngine(cfg, dataDir)
			r, _ := engine.Run(rules, bars)
			report = r
		}()
		if err == nil && report != nil {
			strategy.Version = 1
		}

		// 保存到数据库
		if globalStore != nil {
			_ = globalStore.SaveStrategy(strategy)
			if report != nil {
				report.ID = uuid.New().String()
				report.StrategyID = strategy.ID
				_ = globalStore.SaveBacktest(report)
			}
		}

		c.JSON(http.StatusOK, model.StrategyReport{
			Strategy:  *strategy,
			Backtest: report,
		})
	}
}

// StrategyList 策略列表
func StrategyList() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("userID")
		if userID == "" {
			userID = "default"
		}
		var storeToUse StoreInterface
		if globalStore != nil {
			storeToUse = globalStore
		} else {
			storeToUse = globalStoreInterface
		}
		if storeToUse == nil {
			c.JSON(http.StatusOK, gin.H{"strategies": []model.Strategy{}})
			return
		}
		strategies, err := storeToUse.ListStrategies(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"strategies": strategies})
	}
}

// StrategyGet 策略详情
func StrategyGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var storeToUse StoreInterface
		if globalStore != nil {
			storeToUse = globalStore
		} else {
			storeToUse = globalStoreInterface
		}
		if storeToUse == nil {
			c.JSON(http.StatusOK, gin.H{"strategy": nil})
			return
		}
		strategy, err := storeToUse.GetStrategy(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"strategy": strategy})
	}
}

// StrategyDelete 删除策略
func StrategyDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var storeToUse StoreInterface
		if globalStore != nil {
			storeToUse = globalStore
		} else {
			storeToUse = globalStoreInterface
		}
		if storeToUse == nil {
			c.JSON(http.StatusOK, gin.H{"deleted": true})
			return
		}
		err := storeToUse.DeleteStrategy(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}
