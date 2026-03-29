package handler

import (
	"net/http"
	"quant-agent/pkg/social"

	"github.com/gin-gonic/gin"
)

var (
	eastmoneyClient *social.EastmoneyClient
	xueqiuClient    *social.XueqiuClient
)

func init() {
	eastmoneyClient = social.NewEastmoneyClient()
	xueqiuClient = social.NewXueqiuClient()
}

// GetMarketHot 获取市场热点榜单
// GET /api/v1/social/hot
func GetMarketHot() gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := eastmoneyClient.FetchMarketHot()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": items,
		})
	}
}

// GetStockSentiment 获取个股情绪数据
// GET /api/v1/social/sentiment/:code
func GetStockSentiment() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stock code required"})
			return
		}

		// 并行获取东财和雪球数据
		type result struct {
			Eastmoney *social.SentimentData
			Xueqiu    *social.SentimentData
			Err       error
		}
		ch := make(chan result, 2)

		go func() {
			data, err := eastmoneyClient.FetchStockSentiment(code)
			ch <- result{Eastmoney: data, Err: err}
		}()
		go func() {
			data, err := xueqiuClient.FetchStockSentiment(code)
			ch <- result{Xueqiu: data, Err: err}
		}()

		var emData, xqData *social.SentimentData
		for i := 0; i < 2; i++ {
			r := <-ch
			if r.Eastmoney != nil {
				emData = r.Eastmoney
			}
			if r.Xueqiu != nil {
				xqData = r.Xueqiu
			}
			if r.Err != nil {
				// 只记录错误，不中断
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"eastmoney": emData,
				"xueqiu":    xqData,
			},
		})
	}
}

// GetStockPosts 获取个股讨论帖
// GET /api/v1/social/posts/:code
func GetStockPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stock code required"})
			return
		}

		// 并行获取东财和雪球帖子
		type postResult struct {
			Posts []social.Post
			Err   error
		}
		ch := make(chan postResult, 2)

		go func() {
			posts, err := eastmoneyClient.FetchStockPosts(code)
			ch <- postResult{Posts: posts, Err: err}
		}()
		go func() {
			posts, err := xueqiuClient.FetchStockPosts(code)
			ch <- postResult{Posts: posts, Err: err}
		}()

		var allPosts []social.Post
		for i := 0; i < 2; i++ {
			r := <-ch
			if r.Err == nil {
				allPosts = append(allPosts, r.Posts...)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": allPosts,
		})
	}
}
