package handler

import (
	"fmt"
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

// GetWeChatPosts 搜索微信公众号文章
// GET /api/v1/social/wechat?keyword=贵州茅台&limit=10
func GetWeChatPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyword := c.Query("keyword")
		if keyword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "keyword required"})
			return
		}

		// Parse limit (default 10, max 50)
		limit := 10
		if l := c.Query("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}
		if limit <= 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}

		posts, err := social.FetchWeChatPosts(keyword, limit)
		if err != nil {
			// 不panic，返回空数组
			posts = []social.WeChatPost{}
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": posts,
		})
	}
}

// GetDouyinHot 获取抖音热榜
// GET /api/v1/social/douyin
func GetDouyinHot() gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := social.FetchDouyinHot()
		if err != nil {
			// 不panic，返回空数组
			items = []social.DouyinHotItem{}
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": items,
		})
	}
}

// GetNewsEvents 获取新闻事件
// GET /api/v1/social/news?category=stock
func GetNewsEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.DefaultQuery("category", "stock")
		// 验证category
		validCategories := map[string]bool{"macro": true, "stock": true, "industry": true}
		if !validCategories[category] {
			category = "stock"
		}

		events, err := social.FetchNewsEvents(category)
		if err != nil {
			// 不panic，返回空数组
			events = []social.NewsEvent{}
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": events,
		})
	}
}
