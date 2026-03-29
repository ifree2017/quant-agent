package handler

import (
	"context"
	"net/http"
	"quant-agent/internal/model"
	"quant-agent/internal/optimizer"

	"github.com/gin-gonic/gin"
)

var opt optimizer.Optimizer

// SetOptimizer 设置优化器
func SetOptimizer(o optimizer.Optimizer) {
	opt = o
}

// OptimizeRequest 优化请求
type OptimizeRequest struct {
	Symbol     string                 `json:"symbol"`
	Days       int                    `json:"days"`
	Rules      model.StrategyRules     `json:"rules"`
	ParamSpace map[string][]float64   `json:"paramSpace"`
	Target     string                 `json:"target"`   // sharpe_ratio
	Method     string                 `json:"method"`   // grid | bayesian
}

// Optimize 执行参数优化
func Optimize(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req OptimizeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Days == 0 {
			req.Days = 60
		}
		if req.Target == "" {
			req.Target = "sharpe_ratio"
		}
		if req.Method == "" {
			req.Method = "grid"
		}

		cfg := optimizer.OptimizeConfig{
			Symbol:    req.Symbol,
			DataDir:   dataDir,
			Days:      req.Days,
			StrategyRules: req.Rules,
			ParamSpace: req.ParamSpace,
			Target:    req.Target,
			Method:    req.Method,
			GridMaxCombinations: 500,
		}

		results, err := opt.Optimize(context.Background(), cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}
