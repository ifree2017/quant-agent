package social

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	DouyinPlatform = "douyin"
)

// DouyinHotItem 抖音热榜项
type DouyinHotItem struct {
	Platform  string  `json:"platform"`
	Rank      int     `json:"rank"`
	Title     string  `json:"title"`
	HotValue  int64   `json:"hot_value"`
	Category  string  `json:"category"`
	IsAd      bool    `json:"is_ad"`
	WordCount int     `json:"word_count"` // 热度值（文字描述）
}

// FetchDouyinHot 获取抖音实时热榜
func FetchDouyinHot() ([]DouyinHotItem, error) {
	client := NewClient()

	// 尝试官方API
	items, err := fetchDouyinAPI(client)
	if err == nil && len(items) > 0 {
		return items, nil
	}

	// API失败则尝试页面解析
	return fetchDouyinPage(client)
}

// fetchDouyinAPI 尝试抖音官方API
func fetchDouyinAPI(client *Client) ([]DouyinHotItem, error) {
	apiURL := "https://www.douyin.com/aweme/v1/web/hot/search/list/?device_platform=webapp&aid=6383&channel=channel_pc_web&detail_list=1"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return []DouyinHotItem{}, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.douyin.com/")
	req.Header.Set("Accept", "application/json")

	resp, err := client.DoRequest(req)
	if err != nil {
		return []DouyinHotItem{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []DouyinHotItem{}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []DouyinHotItem{}, nil
	}

	// 尝试解析JSON
	var data struct {
		WordList []struct {
			Word       string `json:"word"`
			HotValue   int64  `json:"hot_value"`
			WordCount  int    `json:"word_count"`
			Category   string `json:"category"`
			IsAd       bool   `json:"is_ad"`
		} `json:"word_list"`
	}

	// 尝试多种可能的JSON结构
	if err := json.Unmarshal(body, &data); err != nil {
		return parseDouyinJSONFallback(body)
	}

	items := make([]DouyinHotItem, 0, len(data.WordList))
	for i, w := range data.WordList {
		items = append(items, DouyinHotItem{
			Platform:  DouyinPlatform,
			Rank:      i + 1,
			Title:     w.Word,
			HotValue:  w.HotValue,
			WordCount: w.WordCount,
			Category:  w.Category,
			IsAd:      w.IsAd,
		})
	}
	return items, nil
}

func parseDouyinJSONFallback(body []byte) ([]DouyinHotItem, error) {
	// 尝试直接解析word_list数组
	var wordList []struct {
		Word      string `json:"word"`
		HotValue  int64  `json:"hot_value"`
		WordCount int    `json:"word_count"`
		Category  string `json:"category"`
		IsAd      bool   `json:"is_ad"`
	}

	if err := json.Unmarshal(body, &wordList); err != nil {
		return []DouyinHotItem{}, nil
	}

	items := make([]DouyinHotItem, 0, len(wordList))
	for i, w := range wordList {
		items = append(items, DouyinHotItem{
			Platform:  DouyinPlatform,
			Rank:      i + 1,
			Title:     w.Word,
			HotValue:  w.HotValue,
			WordCount: w.WordCount,
			Category:  w.Category,
			IsAd:      w.IsAd,
		})
	}
	return items, nil
}

// fetchDouyinPage 尝试解析抖音热榜页面
func fetchDouyinPage(client *Client) ([]DouyinHotItem, error) {
	pageURL := "https://www.douyin.com/hot"

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return []DouyinHotItem{}, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := client.DoRequest(req)
	if err != nil {
		return []DouyinHotItem{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []DouyinHotItem{}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []DouyinHotItem{}, nil
	}

	return parseDouyinPageHTML(string(body))
}

func parseDouyinPageHTML(htmlContent string) ([]DouyinHotItem, error) {
	// 尝试从页面HTML中提取JSON数据
	// 抖音页面通常将数据嵌入在script标签中
	re := regexp.MustCompile(`"hot_list":\s*(\[.*?\])`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) >= 2 {
		var data []struct {
			Word      string `json:"word"`
			HotValue  int64  `json:"hot_value"`
			WordCount int    `json:"word_count"`
			Category  string `json:"category"`
		}
		if err := json.Unmarshal([]byte(matches[1]), &data); err == nil {
			items := make([]DouyinHotItem, 0, len(data))
			for i, w := range data {
				items = append(items, DouyinHotItem{
					Platform:  DouyinPlatform,
					Rank:      i + 1,
					Title:     w.Word,
					HotValue:  w.HotValue,
					WordCount: w.WordCount,
					Category:  w.Category,
					IsAd:      false,
				})
			}
			if len(items) > 0 {
				return items, nil
			}
		}
	}

	// 尝试HTML解析方式
	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return []DouyinHotItem{}, nil
	}

	// 尝试多种选择器
	nodes := htmlquery.Find(doc, "//div[@class='hot-list-item'] | //li[@class='hot-item'] | //div[@data-e2e='hot-list-item']")
	if len(nodes) == 0 {
		return []DouyinHotItem{}, nil
	}

	items := make([]DouyinHotItem, 0, len(nodes))
	for i, node := range nodes {
		title := extractDouyinText(node, "//div[@class='title']//a | //h3 | //span[@class='title']")
		hotValueStr := extractDouyinText(node, "//span[@class='hot-value'] | //div[@class='hot-value']")
		category := extractDouyinText(node, "//span[@class='category'] | //div[@class='category']")

		hotValue := parseHotValue(hotValueStr)

		if title != "" {
			items = append(items, DouyinHotItem{
				Platform:  DouyinPlatform,
				Rank:      i + 1,
				Title:     cleanText(title),
				HotValue:  hotValue,
				WordCount: int(hotValue),
				Category:  cleanText(category),
				IsAd:      false,
			})
		}
	}

	return items, nil
}

func extractDouyinText(node *html.Node, xpath string) string {
	n := htmlquery.FindOne(node, xpath)
	if n == nil {
		return ""
	}
	return htmlquery.InnerText(n)
}

func parseHotValue(s string) int64 {
	s = strings.TrimSpace(s)
	// 处理 "1234万" 格式
	if strings.HasSuffix(s, "万") {
		var val float64
		fmt.Sscanf(s[:len(s)-1], "%f", &val)
		return int64(val * 10000)
	}
	// 处理纯数字
	var val int64
	fmt.Sscanf(s, "%d", &val)
	return val
}
