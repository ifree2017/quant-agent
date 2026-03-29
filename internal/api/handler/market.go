package handler

import (
	"net/http"
	"quant-agent/internal/data"

	"github.com/gin-gonic/gin"
)

// GetMarketInfo 获取市场信息
func GetMarketInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Query("symbol")
		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol required"})
			return
		}

		loader := data.NewLoaderV2("./data")
		bars, market, err := loader.LoadBarsAdvanced(symbol, 1)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"symbol": symbol,
			"market": market,
			"count":  len(bars),
		})
	}
}
