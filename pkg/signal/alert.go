package signal

import (
	"context"
	"strings"
	"sync"
	"time"

	"quant-agent/pkg/social"
)

// DetectAlerts 检测舆情预警（负面新闻）
// 逻辑：
//  1. 获取个股相关新闻
//  2. 检测负面关键词密度（黑天鹅、减持、解禁、业绩下滑、监管）
//  3. 3条连续负面 → 高危预警
//  4. 输出：预警等级(LOW/MEDIUM/HIGH/CRITICAL)、描述、持续时间
func (e *Engine) DetectAlerts(stockCode string) ([]AlertSignal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return e.DetectAlertsWithContext(ctx, stockCode)
}

// DetectAlertsWithContext 带超时的舆情预警检测
func (e *Engine) DetectAlertsWithContext(ctx context.Context, stockCode string) ([]AlertSignal, error) {
	type newsResult struct {
		platform string
		news     interface{}
		err      error
	}

	resultCh := make(chan newsResult, 3)
	var wg sync.WaitGroup

	// 并发获取各平台新闻
	wg.Add(3)

	// 东财帖子
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			resultCh <- newsResult{err: ctx.Err()}
			return
		default:
		}

		client := social.NewEastmoneyClient()
		posts, err := client.FetchStockPosts(stockCode)
		resultCh <- newsResult{platform: "eastmoney_posts", news: posts, err: err}
	}()

	// 雪球帖子
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			resultCh <- newsResult{err: ctx.Err()}
			return
		default:
		}

		client := social.NewXueqiuClient()
		posts, err := client.FetchStockPosts(stockCode)
		resultCh <- newsResult{platform: "xueqiu_posts", news: posts, err: err}
	}()

	// 新闻事件
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			resultCh <- newsResult{err: ctx.Err()}
			return
		default:
		}

		events, err := social.FetchNewsEvents("stock")
		resultCh <- newsResult{platform: "news", news: events, err: err}
	}()

	// 等待完成
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 收集并分析负面关键词
	negativeMatches := make([]negativeMatch, 0)
	newsCount := 0

	for result := range resultCh {
		if result.err != nil {
			continue
		}

		switch result.platform {
		case "eastmoney_posts":
			if posts, ok := result.news.([]social.Post); ok {
				for _, post := range posts {
					newsCount++
					matches := e.analyzeNegativeKeywords(post.Title + " " + post.Content)
					negativeMatches = append(negativeMatches, matches...)
				}
			}
		case "xueqiu_posts":
			if posts, ok := result.news.([]social.Post); ok {
				for _, post := range posts {
					newsCount++
					matches := e.analyzeNegativeKeywords(post.Title + " " + post.Content)
					negativeMatches = append(negativeMatches, matches...)
				}
			}
		case "news":
			if events, ok := result.news.([]social.NewsEvent); ok {
				for _, event := range events {
					newsCount++
					matches := e.analyzeNegativeKeywords(event.Title + " " + event.Summary)
					negativeMatches = append(negativeMatches, matches...)
				}
			}
		}
	}

	// 生成预警信号
	return e.generateAlerts(stockCode, negativeMatches, newsCount), nil
}

// negativeMatch 负面关键词匹配记录
type negativeMatch struct {
	keyword   string
	text      string
	timestamp time.Time
}

// analyzeNegativeKeywords 分析文本中的负面关键词
func (e *Engine) analyzeNegativeKeywords(text string) []negativeMatch {
	matches := make([]negativeMatch, 0)
	textLower := strings.ToLower(text)

	for _, kw := range NegativeKeywords {
		if strings.Contains(textLower, strings.ToLower(kw)) {
			matches = append(matches, negativeMatch{
				keyword:   kw,
				text:      text,
				timestamp: time.Now(),
			})
		}
	}

	return matches
}

// generateAlerts 生成预警信号
func (e *Engine) generateAlerts(stockCode string, matches []negativeMatch, newsCount int) []AlertSignal {
	if len(matches) == 0 {
		return []AlertSignal{}
	}

	// 按关键词分组统计
	keywordCount := make(map[string]int)
	for _, m := range matches {
		keywordCount[m.keyword]++
	}

	// 计算预警等级
	alertLevel := e.calcAlertLevel(matches, newsCount)

	// 生成描述
	description := e.generateAlertDescription(keywordCount, matches)

	// 确定持续时间
	duration := e.determineDuration(matches)

	// 提取最常见的关键词
	topKeywords := e.getTopKeywords(keywordCount, 5)

	alert := AlertSignal{
		StockCode:   stockCode,
		AlertLevel:  alertLevel,
		Title:       e.generateAlertTitle(alertLevel, stockCode),
		Description: description,
		Keywords:    topKeywords,
		NewsCount:   newsCount,
		Duration:    duration,
		Source:      "news_alert",
		CreatedAt:   time.Now(),
	}

	return []AlertSignal{alert}
}

// calcAlertLevel 计算预警等级
func (e *Engine) calcAlertLevel(matches []negativeMatch, newsCount int) AlertLevel {
	totalNegative := len(matches)
	uniqueKeywords := len(e.uniqueKeywords(matches))

	// 3条连续负面 → CRITICAL
	if totalNegative >= 3 && uniqueKeywords >= 2 {
		return AlertCritical
	}

	// 多个负面关键词 → HIGH
	if totalNegative >= 2 && uniqueKeywords >= 2 {
		return AlertHigh
	}

	// 1个明显负面关键词 → MEDIUM
	if totalNegative >= 1 && uniqueKeywords >= 1 {
		return AlertMedium
	}

	return AlertLow
}

// uniqueKeywords 提取唯一关键词
func (e *Engine) uniqueKeywords(matches []negativeMatch) []string {
	seen := make(map[string]bool)
	out := make([]string, 0)
	for _, m := range matches {
		if !seen[m.keyword] {
			seen[m.keyword] = true
			out = append(out, m.keyword)
		}
	}
	return out
}

// generateAlertDescription 生成预警描述
func (e *Engine) generateAlertDescription(keywordCount map[string]int, matches []negativeMatch) string {
	if len(keywordCount) == 0 {
		return "未检测到明显负面信息"
	}

	// 按频率排序
	sorted := e.sortByCount(keywordCount)
	top3 := sorted
	if len(top3) > 3 {
		top3 = top3[:3]
	}

	var parts []string
	for _, item := range top3 {
		if item.count > 1 {
			parts = append(parts, item.keyword)
		}
	}

	if len(parts) > 0 {
		return "检测到负面关键词：" + strings.Join(parts, "、")
	}

	return "检测到轻度负面信号，建议关注"
}

// sortByCount 按计数排序
func (e *Engine) sortByCount(m map[string]int) []struct {
	keyword string
	count   int
} {
	sorted := make([]struct {
		keyword string
		count   int
	}, 0, len(m))
	for k, v := range m {
		sorted = append(sorted, struct {
			keyword string
			count   int
		}{keyword: k, count: v})
	}

	// 冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// generateAlertTitle 生成预警标题
func (e *Engine) generateAlertTitle(level AlertLevel, stockCode string) string {
	switch level {
	case AlertCritical:
		return "🚨 [" + stockCode + "] 高危预警：检测到多重负面信号"
	case AlertHigh:
		return "⚠️ [" + stockCode + "] 高风险预警"
	case AlertMedium:
		return "📊 [" + stockCode + "] 中度风险提示"
	case AlertLow:
		return "📋 [" + stockCode + "] 低风险关注"
	default:
		return "📋 [" + stockCode + "] 风险提示"
	}
}

// determineDuration 判断持续时间
func (e *Engine) determineDuration(matches []negativeMatch) string {
	count := len(matches)

	if count >= 5 {
		return "中期"
	}
	if count >= 3 {
		return "短期"
	}
	if count >= 1 {
		return "临时"
	}
	return "待观察"
}

// getTopKeywords 获取高频关键词
func (e *Engine) getTopKeywords(keywordCount map[string]int, limit int) []string {
	sorted := e.sortByCount(keywordCount)

	top := make([]string, 0, limit)
	for i, item := range sorted {
		if i >= limit {
			break
		}
		top = append(top, item.keyword)
	}

	return top
}
