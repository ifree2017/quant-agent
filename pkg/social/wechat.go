package social

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	WechatPlatform = "wechat"
)

// WeChatPost 搜狗微信文章
type WeChatPost struct {
	Platform    string    `json:"platform"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Source      string    `json:"source"` // 公众号名称
	PublishTime time.Time `json:"publish_time"`
	URL         string    `json:"url"`
}

// FetchWeChatPosts 搜索微信公众号文章
func FetchWeChatPosts(keyword string, limit int) ([]WeChatPost, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	client := NewClient()
	apiURL := fmt.Sprintf(
		"https://weixin.sogou.com/weixin?type=2&query=%s&ie=utf8&s_from=input&_sug_=n&_sug_type_=",
		url.QueryEscape(keyword),
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://weixin.sogou.com/")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.DoRequest(req)
	if err != nil {
		return []WeChatPost{}, nil // 不panic，返回空数组
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []WeChatPost{}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []WeChatPost{}, nil
	}

	return parseWeChatHTML(string(body), limit)
}

func parseWeChatHTML(htmlContent string, limit int) ([]WeChatPost, error) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return []WeChatPost{}, nil
	}

	// 查找所有文章条目
	nodes := htmlquery.Find(doc, "//div[@class='news-box']//li[@class='news-list']")
	if len(nodes) == 0 {
		// 尝试备用选择器
		nodes = htmlquery.Find(doc, "//div[@class='news-list']//li")
	}
	if len(nodes) == 0 {
		nodes = htmlquery.Find(doc, "//li[@class='news-wrap']")
	}

	posts := make([]WeChatPost, 0, len(nodes))
	for i, node := range nodes {
		if i >= limit {
			break
		}

		title := extractWeChatText(node, "//h3[@class='tit']//a | //h3/a | //a[@class='news-title']")
		if title == "" {
			title = extractWeChatText(node, "//*[@class='title']//a | //a[@class='news-title']")
		}

		summary := extractWeChatText(node, "//p[@class='txt-info'] | //p[@class='description']")
		if summary == "" {
			summary = extractWeChatText(node, "//div[@class='news-desc']")
		}

		source := extractWeChatText(node, "//div[@class='account']//a | //span[@class='account']//a | //*[@class='s-p']")
		if source == "" {
			source = extractWeChatText(node, "//div[@class='s-p']")
		}

		timeStr := extractWeChatText(node, "//span[@class='s2'] | //div[@class='news-desc']//span")
		pt := parseWeChatTime(timeStr)

		href := extractWeChatAttr(node, "//h3[@class='tit']//a | //h3/a | //a[@class='news-title']", "href")
		if href == "" {
			href = extractWeChatAttr(node, "//*[@class='title']//a", "href")
		}

		if title != "" {
			posts = append(posts, WeChatPost{
				Platform:    WechatPlatform,
				Title:       cleanText(title),
				Summary:     cleanText(summary),
				Source:      cleanText(source),
				PublishTime: pt,
				URL:         href,
			})
		}
	}

	// 如果解析失败，返回空数组不panic
	return posts, nil
}

func extractWeChatText(node *html.Node, xpath string) string {
	n := htmlquery.FindOne(node, xpath)
	if n == nil {
		return ""
	}
	return htmlquery.InnerText(n)
}

func extractWeChatAttr(node *html.Node, xpath string, attr string) string {
	n := htmlquery.FindOne(node, xpath)
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}

func parseWeChatTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}

	// 尝试多种时间格式
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02",
		"01-02 15:04",
		"MM-dd hh:mm",
	}
	for _, fmt := range formats {
		if t, err := time.ParseInLocation(fmt, s, time.Local); err == nil {
			return t
		}
	}

	// 相对时间
	if strings.Contains(s, "前") || strings.Contains(s, "天") || strings.Contains(s, "小时") {
		return time.Now()
	}

	return time.Time{}
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return s
}
