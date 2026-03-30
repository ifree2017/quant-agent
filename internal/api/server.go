package api

import (
	"log"
	"os"
	"quant-agent/internal/ai"
	"quant-agent/internal/api/handler"
	"quant-agent/internal/store"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务器
type Server struct {
	engine   *gin.Engine
	dataDir  string
	aiClient *ai.Client
}

// NewServer 创建服务器
func NewServer(dataDir string, llmURL, llmToken string, store *store.Store) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	s := &Server{
		engine:   r,
		dataDir:  dataDir,
		aiClient: ai.NewClient(llmURL, llmToken),
	}

	// 注入 store
	handler.SetStore(store)
	handler.SetBacktestStore(store)

	s.routes()
	return s
}

func (s *Server) routes() {
	// 健康检查
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 风格分析
	s.engine.POST("/api/style/analyze", handler.StyleAnalyze(s.aiClient))

	// 策略生成
	s.engine.POST("/api/strategy/generate", handler.StrategyGenerate(s.aiClient, s.dataDir))

	// 策略列表
	s.engine.GET("/api/strategies", handler.StrategyList())

	// 策略详情
	s.engine.GET("/api/strategies/:id", handler.StrategyGet())

	// 删除策略
	s.engine.DELETE("/api/strategies/:id", handler.StrategyDelete())

	// 执行回测
	s.engine.POST("/api/backtest", handler.BacktestRun(s.dataDir))

	// 回测结果
	s.engine.GET("/api/backtest/:id", handler.BacktestGet())

	// 参数优化
	s.engine.POST("/api/optimize", handler.Optimize(s.dataDir))

	// 市场信息
	s.engine.GET("/api/market/info", handler.GetMarketInfo())

	// 社交情绪数据
	s.engine.GET("/api/v1/social/hot", handler.GetMarketHot())
	s.engine.GET("/api/v1/social/sentiment/:code", handler.GetStockSentiment())
	s.engine.GET("/api/v1/social/posts/:code", handler.GetStockPosts())
	s.engine.GET("/api/v1/social/wechat", handler.GetWeChatPosts())
	s.engine.GET("/api/v1/social/douyin", handler.GetDouyinHot())
	s.engine.GET("/api/v1/social/news", handler.GetNewsEvents())

	// 社交情绪信号
	s.engine.GET("/api/v1/signal/:code", handler.GetSignal())
	s.engine.GET("/api/v1/signal/sector", handler.GetSectorSignals())
	s.engine.GET("/api/v1/signal/alerts", handler.GetAlerts())

	// 出货识别 & 买点识别（F12/F13）
	s.engine.GET("/api/v1/distribution/:code", handler.GetDistribution())
	s.engine.GET("/api/v1/buypoint/:code", handler.GetBuyPoint())
}

// Run 启动服务器
func (s *Server) Run(addr string) {
	log.Printf("QuantAgent server starting on %s", addr)
	if err := s.engine.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
