package social

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	EastmoneyPlatform = "eastmoney"
)

// EastmoneyClient 东财数据客户端
type EastmoneyClient struct {
	client *Client
}

// NewEastmoneyClient 创建东财客户端
func NewEastmoneyClient() *EastmoneyClient {
	return &EastmoneyClient{
		client: NewClient(),
	}
}

// FetchMarketHot 获取市场热点板块/个股
func (c *EastmoneyClient) FetchMarketHot() ([]MarketHotItem, error) {
	ts := time.Now().UnixMilli()
	apiURL := fmt.Sprintf(
		"https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=20&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:13,m:1+t:2,m:1+t:23&fields=f12,f14,f3,f62&_=%d",
		ts,
	)

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("eastmoney FetchMarketHot request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("eastmoney API status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	var data struct {
		Data struct {
			Diff []EastmoneyItem `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return c.parseMarketHotFallback(body)
	}

	items := make([]MarketHotItem, 0, len(data.Data.Diff))
	for i, item := range data.Data.Diff {
		heat := calcHeatScore(item.F3, item.F62)
		items = append(items, MarketHotItem{
			Rank:      i + 1,
			StockCode: item.F12,
			StockName: item.F14,
			ChangePct: item.F3,
			HeatScore: heat,
			Platform:  EastmoneyPlatform,
		})
	}
	return items, nil
}

func (c *EastmoneyClient) parseMarketHotFallback(body []byte) ([]MarketHotItem, error) {
	var diffs []EastmoneyItem
	if err := json.Unmarshal(body, &diffs); err != nil {
		return nil, fmt.Errorf("json parse failed, response: %s", string(body[:min(200, len(body))]))
	}
	items := make([]MarketHotItem, 0, len(diffs))
	for i, item := range diffs {
		heat := calcHeatScore(item.F3, item.F62)
		items = append(items, MarketHotItem{
			Rank:      i + 1,
			StockCode: item.F12,
			StockName: item.F14,
			ChangePct: item.F3,
			HeatScore: heat,
			Platform:  EastmoneyPlatform,
		})
	}
	return items, nil
}

// FetchStockPosts 获取个股帖子/评论热度
func (c *EastmoneyClient) FetchStockPosts(stockCode string) ([]Post, error) {
	ts := time.Now().UnixMilli()
	apiURL := fmt.Sprintf(
		"https://guba.eastmoney.com/interface/GetData.aspx?cb=jQuery&appid=GetNewsColumn&param=%%7B%%22column%%22%%3A%%22stock%%22%%2C%%22pkColumn%%22%%3A%%22%s%%22%%2C%%22pageIndex%%22%%3A1%%2C%%22pageSize%%22%%3A20%%2C%%22sortType%%22%%3A1%%7D&_=%d",
		url.QueryEscape(stockCode), ts,
	)

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("eastmoney FetchStockPosts request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("eastmoney guba API status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	return c.parsePostsJSONP(string(body), stockCode)
}

func (c *EastmoneyClient) parsePostsJSONP(raw string, stockCode string) ([]Post, error) {
	re := regexp.MustCompile(`jQuery[^(]*\((.*)\)`)
	matches := re.FindStringSubmatch(raw)
	if len(matches) < 2 {
		return c.parsePostsDirect([]byte(raw), stockCode)
	}

	jsonStr := matches[1]
	var gubaResp struct {
		Items []struct {
			Title      string `json:"title"`
			Content    string `json:"content"`
			AuthorName string `json:"authorname"`
			LikeCount  int    `json:"like_count"`
			ReplyCount int    `json:"reply_count"`
			ShowTime   string `json:"show_time"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &gubaResp); err != nil {
		return c.parsePostsDirect([]byte(raw), stockCode)
	}

	posts := make([]Post, 0, len(gubaResp.Items))
	for _, item := range gubaResp.Items {
		keywords := extractKeywords(item.Title + " " + item.Content)
		sentiment := calcSentiment(item.Title, item.Content, item.LikeCount, item.ReplyCount, 0)
		pt, _ := parseEastmoneyTime(item.ShowTime)
		posts = append(posts, Post{
			StockCode:   stockCode,
			Platform:    EastmoneyPlatform,
			Title:       item.Title,
			Content:     item.Content,
			Author:      item.AuthorName,
			Likes:       item.LikeCount,
			Comments:    item.ReplyCount,
			Sentiment:   sentiment,
			PublishTime: pt,
			FetchedAt:   time.Now(),
			Keywords:    keywords,
		})
	}
	return posts, nil
}

func (c *EastmoneyClient) parsePostsDirect(body []byte, stockCode string) ([]Post, error) {
	return []Post{}, nil
}

// FetchStockSentiment 获取个股情绪指标
func (c *EastmoneyClient) FetchStockSentiment(stockCode string) (*SentimentData, error) {
	ts := time.Now().UnixMilli()
	apiURL := fmt.Sprintf(
		"https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=5&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:13,m:1+t:2,m:1+t:23&fields=f12,f14,f3,f62&_=%d",
		ts,
	)

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("eastmoney FetchStockSentiment request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Data struct {
			Diff []EastmoneyItem `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var found *EastmoneyItem
	for i := range data.Data.Diff {
		if data.Data.Diff[i].F12 == stockCode {
			found = &data.Data.Diff[i]
			break
		}
	}

	if found == nil {
		return &SentimentData{
			StockCode:      stockCode,
			Platform:       EastmoneyPlatform,
			SentimentScore: 0,
			PostCount:      0,
			CommentCount:   0,
			HeatScore:      0,
			Keywords:       []string{},
			FetchedAt:      time.Now(),
		}, nil
	}

	score := calcMarketSentiment(found.F3, found.F62)
	return &SentimentData{
		StockCode:      stockCode,
		Platform:       EastmoneyPlatform,
		SentimentScore: score,
		HeatScore:     calcHeatScore(found.F3, found.F62),
		PostCount:     0,
		CommentCount:  0,
		Keywords:      []string{},
		FetchedAt:     time.Now(),
	}, nil
}

func calcHeatScore(changePct, netInflow float64) float64 {
	changeScore := math.Abs(changePct) * 10
	inflowScore := math.Abs(netInflow) / 1e8
	return math.Round((changeScore+inflowScore)*100) / 100
}

func calcMarketSentiment(changePct, netInflow float64) float64 {
	score := changePct * 2
	if netInflow > 0 {
		score += math.Min(netInflow/1e8*5, 30)
	} else {
		score += math.Max(netInflow/1e8*5, -30)
	}
	score = math.Max(-100, math.Min(100, score))
	return math.Round(score*100) / 100
}

func parseEastmoneyTime(s string) (time.Time, error) {
	if strings.Contains(s, "前") {
		return time.Now(), nil
	}
	return time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
}
