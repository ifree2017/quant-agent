package social

import "time"

// SentimentScore 情绪评分 (-100 极度恐慌 ~ +100 极度贪婪)
type SentimentScore float64

// Post 帖子数据
type Post struct {
	ID          int64     `json:"id"`
	StockCode   string    `json:"stock_code"`
	Platform    string    `json:"platform"` // eastmoney | xueqiu
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	Likes       int       `json:"likes"`
	Comments    int       `json:"comments"`
	Sentiment   float64   `json:"sentiment"`
	PublishTime time.Time `json:"publish_time"`
	FetchedAt   time.Time `json:"fetched_at"`
	Keywords    []string  `json:"keywords"`
}

// SentimentData 情绪数据
type SentimentData struct {
	StockCode     string    `json:"stock_code"`
	Platform      string    `json:"platform"`
	SentimentScore float64  `json:"sentiment_score"` // -100 ~ +100
	PostCount     int       `json:"post_count"`
	CommentCount  int       `json:"comment_count"`
	HeatScore     float64   `json:"heat_score"`
	Keywords      []string  `json:"keywords"`
	FetchedAt     time.Time `json:"fetched_at"`
}

// MarketHotItem 市场热点项
type MarketHotItem struct {
	Rank       int     `json:"rank"`
	StockCode  string  `json:"stock_code"`
	StockName  string  `json:"stock_name"`
	ChangePct  float64 `json:"change_pct"` // 涨跌幅 %
	HeatScore  float64 `json:"heat_score"`
	Platform   string  `json:"platform"`
}

// EastmoneyResponse 东财API通用响应
type EastmoneyResponse struct {
	Data EastmoneyData `json:"data"`
}

type EastmoneyData struct {
	Diff []EastmoneyItem `json:"diff"`
}

type EastmoneyItem struct {
	F12 string  `json:"f12"` // 股票代码
	F14 string  `json:"f14"` // 股票名称
	F3  float64 `json:"f3"`  // 涨跌幅
	F62 float64 `json:"f62"` // 主力净流入
}
