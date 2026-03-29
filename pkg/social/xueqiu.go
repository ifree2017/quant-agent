package social

import (
	"fmt"
	"time"
)

const (
	XueqiuPlatform = "xueqiu"
)

// XueqiuClient 雪球数据客户端（模拟数据）
// 雪球反爬严格，线上环境需要真实cookie，此处返回结构化模拟数据
type XueqiuClient struct {
	client *Client
}

// NewXueqiuClient 创建雪球客户端
func NewXueqiuClient() *XueqiuClient {
	return &XueqiuClient{
		client: NewClient(),
	}
}

// FetchStockPosts 获取雪球讨论（模拟数据）
func (c *XueqiuClient) FetchStockPosts(stockCode string) ([]Post, error) {
	// 雪球反爬严格，返回结构化模拟数据
	// 实际生产环境需要维护真实cookie和UA
	now := time.Now()
	posts := []Post{
		{
			StockCode:   stockCode,
			Platform:    XueqiuPlatform,
			Title:       fmt.Sprintf("$%s$ 看好这只股票的三个理由", stockCode),
			Content:     "估值处于历史低位，基本面优秀，行业龙头地位稳固。近期成交量放大，主力资金持续流入，技术面呈现多头排列。",
			Author:      "价值投资者小明",
			Likes:       328,
			Comments:    45,
			Sentiment:   72.5,
			PublishTime: now.Add(-2 * time.Hour),
			FetchedAt:   now,
			Keywords:    []string{"看好", "低估", "龙头", "资金流入"},
		},
		{
			StockCode:   stockCode,
			Platform:    XueqiuPlatform,
			Title:       fmt.Sprintf("$%s$ 短期风险提示", stockCode),
			Content:     "高位横盘多日，MACD出现死叉，短期可能回调。建议设好止损位，不宜追高。关注下方支撑位。",
			Author:      "技术分析爱好者",
			Likes:       156,
			Comments:    28,
			Sentiment:   -35.0,
			PublishTime: now.Add(-5 * time.Hour),
			FetchedAt:   now,
			Keywords:    []string{"止损", "回调", "MACD", "死叉"},
		},
		{
			StockCode:   stockCode,
			Platform:    XueqiuPlatform,
			Title:       fmt.Sprintf("$%s$ 抄底机会来了？", stockCode),
			Content:     "今日大幅下跌，但从基本面看并无实质性利空。恐慌情绪主导盘面，逆向思维来看可能是布局良机。",
			Author:      "逆向投资老王",
			Likes:       412,
			Comments:    67,
			Sentiment:   55.0,
			PublishTime: now.Add(-8 * time.Hour),
			FetchedAt:   now,
			Keywords:    []string{"抄底", "恐慌", "逆向", "布局"},
		},
		{
			StockCode:   stockCode,
			Platform:    XueqiuPlatform,
			Title:       fmt.Sprintf("$%s$ 财报解读：业绩超预期", stockCode),
			Content:     "Q3财报发布，营收和利润均超市场预期。毛利率提升2个百分点，现金流充裕。管理层上调全年指引。",
			Author:      "基本面研究猿",
			Likes:       289,
			Comments:    53,
			Sentiment:   68.0,
			PublishTime: now.Add(-12 * time.Hour),
			FetchedAt:   now,
			Keywords:    []string{"财报", "超预期", "营收", "利润"},
		},
		{
			StockCode:   stockCode,
			Platform:    XueqiuPlatform,
			Title:       fmt.Sprintf("$%s$ 逃顶还是持有？", stockCode),
			Content:     "涨了这么多，到底是走还是留？我的操作是分批减仓，锁住利润。宁可少赚，不要被套。",
			Author:      "稳健理财者",
			Likes:       198,
			Comments:    34,
			Sentiment:   15.0,
			PublishTime: now.Add(-18 * time.Hour),
			FetchedAt:   now,
			Keywords:    []string{"逃顶", "减仓", "止盈", "锁利"},
		},
	}
	return posts, nil
}

// FetchStockSentiment 获取雪球情绪（模拟数据）
func (c *XueqiuClient) FetchStockSentiment(stockCode string) (*SentimentData, error) {
	posts, err := c.FetchStockPosts(stockCode)
	if err != nil {
		return nil, err
	}

	var totalSentiment float64
	keywordsCount := make(map[string]int)
	for _, p := range posts {
		totalSentiment += p.Sentiment
		for _, k := range p.Keywords {
			keywordsCount[k]++
		}
	}

	avgSentiment := 0.0
	if len(posts) > 0 {
		avgSentiment = totalSentiment / float64(len(posts))
	}

	// 取前5个高频关键词
	topKeywords := make([]string, 0, 5)
	for k, v := range keywordsCount {
		if v >= 2 {
			topKeywords = append(topKeywords, k)
		}
	}

	return &SentimentData{
		StockCode:      stockCode,
		Platform:       XueqiuPlatform,
		SentimentScore: round2(avgSentiment),
		PostCount:      len(posts),
		CommentCount:   sumComments(posts),
		HeatScore:      calcHeat(posts),
		Keywords:       topKeywords,
		FetchedAt:      time.Now(),
	}, nil
}

func sumComments(posts []Post) int {
	total := 0
	for _, p := range posts {
		total += p.Comments
	}
	return total
}

func calcHeat(posts []Post) float64 {
	var heat float64
	for _, p := range posts {
		heat += float64(p.Likes)*1.0 + float64(p.Comments)*2.0
	}
	return round2(heat / float64(len(posts)))
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
