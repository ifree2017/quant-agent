package social

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	NewsPlatform = "news"
)

// NewsEvent 新闻事件
type NewsEvent struct {
	Platform    string    `json:"platform"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Source      string    `json:"source"`
	PublishTime time.Time `json:"publish_time"`
	URL         string    `json:"url"`
	Category    string    `json:"category"` // macro | stock | industry
}

// FetchNewsEvents 获取财经新闻
// category: "macro"(宏观) | "stock"(个股) | "industry"(行业)
func FetchNewsEvents(category string) ([]NewsEvent, error) {
	validCategories := map[string]bool{"macro": true, "stock": true, "industry": true}
	if !validCategories[category] {
		category = "stock"
	}

	// 尝试多个新闻源
	sources := []struct {
		name     string
		fetchFn  func() ([]NewsEvent, error)
		priority int
	}{
		{"tencent", fetchTencentNews, 1},
		{"eastmoney", func() ([]NewsEvent, error) { return fetchEastmoneyNews(category) }, 2},
	}

	for _, src := range sources {
		events, err := src.fetchFn()
		if err == nil && len(events) > 0 {
			// 添加category标记
			for i := range events {
				if events[i].Category == "" {
					events[i].Category = category
				}
			}
			return events, nil
		}
	}

	return []NewsEvent{}, nil
}

// fetchTencentNews 获取腾讯新闻
func fetchTencentNews() ([]NewsEvent, error) {
	client := NewClient()
	ts := time.Now().UnixMilli()
	apiURL := fmt.Sprintf(
		"https://view.inews.qq.com/g2/getOnsInfo?name=dunhuang_news&callback=callback&_=%d",
		ts,
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return []NewsEvent{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://news.qq.com/")

	resp, err := client.DoRequest(req)
	if err != nil {
		return []NewsEvent{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []NewsEvent{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []NewsEvent{}, err
	}

	return parseTencentNews(string(body))
}

func parseTencentNews(raw string) ([]NewsEvent, error) {
	// JSONP回调格式: callback({...})
	re := regexp.MustCompile(`callback\((.*)\)`)
	matches := re.FindStringSubmatch(raw)
	if len(matches) < 2 {
		// 尝试直接解析JSON
		return parseTencentNewsDirect([]byte(raw))
	}

	var data struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(matches[1]), &data); err != nil {
		return parseTencentNewsDirect([]byte(raw))
	}

	var newsList []struct {
		Title    string `json:"title"`
		Summary  string `json:"summary"`
		Source   string `json:"source"`
		PubTime  string `json:"pub_time"`
		URL      string `json:"url"`
	}
	if err := json.Unmarshal([]byte(data.Data), &newsList); err != nil {
		return []NewsEvent{}, nil
	}

	return convertTencentNews(newsList)
}

func parseTencentNewsDirect(body []byte) ([]NewsEvent, error) {
	// 尝试直接解析
	var data struct {
		Data []struct {
			Title   string `json:"title"`
			Summary string `json:"summary"`
			Source  string `json:"source"`
			PubTime string `json:"pub_time"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return []NewsEvent{}, nil
	}

	newsList := make([]struct {
		Title    string `json:"title"`
		Summary  string `json:"summary"`
		Source   string `json:"source"`
		PubTime  string `json:"pub_time"`
		URL      string `json:"url"`
	}, len(data.Data))
	for i, item := range data.Data {
		newsList[i] = item
	}

	return convertTencentNews(newsList)
}

func convertTencentNews(newsList []struct {
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	PubTime  string `json:"pub_time"`
	URL      string `json:"url"`
}) ([]NewsEvent, error) {
	events := make([]NewsEvent, 0, len(newsList))
	for _, item := range newsList {
		pt := parseNewsTime(item.PubTime)
		events = append(events, NewsEvent{
			Platform:    NewsPlatform,
			Title:       item.Title,
			Summary:     item.Summary,
			Source:      item.Source,
			PublishTime: pt,
			URL:         item.URL,
			Category:    "stock",
		})
	}
	return events, nil
}

// fetchEastmoneyNews 获取东财新闻
func fetchEastmoneyNews(category string) ([]NewsEvent, error) {
	client := NewClient()

	// 东财早知道/午间回顾 API
	apiURL := "https://np-anotice-stock.eastmoney.com/api/security/ann?sr=-1&page_size=20&page_index=1&ann_type=A&client_source=web"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return []NewsEvent{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.eastmoney.com/")

	resp, err := client.DoRequest(req)
	if err != nil {
		return []NewsEvent{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []NewsEvent{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []NewsEvent{}, err
	}

	return parseEastmoneyNews(string(body), category)
}

func parseEastmoneyNews(raw string, category string) ([]NewsEvent, error) {
	var data struct {
		Data []struct {
			Title       string `json:"title"`
			NoticeDate  string `json:"notice_date"`
			Source      string `json:"source"`
			 securities []struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"securities"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return []NewsEvent{}, nil
	}

	events := make([]NewsEvent, 0, len(data.Data))
	for _, item := range data.Data {
		pt := parseNewsTime(item.NoticeDate)

		// 拼接关联股票
		var stockInfo string
		for _, s := range item.securities {
			if stockInfo != "" {
				stockInfo += ", "
			}
			stockInfo += s.Name + "(" + s.Code + ")"
		}

		summary := item.Source
		if stockInfo != "" {
			summary = summary + " | 关联: " + stockInfo
		}

		events = append(events, NewsEvent{
			Platform:    NewsPlatform,
			Title:       item.Title,
			Summary:     summary,
			Source:      item.Source,
			PublishTime: pt,
			URL:         "",
			Category:    category,
		})
	}

	return events, nil
}

func parseNewsTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}

	// 尝试多种格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04",
		"2006-01-02",
		"01-02 15:04:05",
	}

	for _, fmt := range formats {
		if t, err := time.ParseInLocation(fmt, s, time.Local); err == nil {
			return t
		}
	}

	return time.Time{}
}
