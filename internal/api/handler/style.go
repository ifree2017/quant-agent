package handler

import (
	"net/http"
	"quant-agent/internal/ai"
	"quant-agent/internal/model"

	"github.com/gin-gonic/gin"
)

// StyleAnalyze 风格分析
func StyleAnalyze(aiClient *ai.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.StyleAnalyzeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		profile, err := aiClient.AnalyzeStyle(req.Records)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		profile.UserID = req.UserID
		c.JSON(http.StatusOK, &profile)
	}
}
