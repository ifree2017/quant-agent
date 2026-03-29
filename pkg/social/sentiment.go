package social

import (
	"regexp"
	"strings"
)

// 情绪关键词
var (
	bullishKeywords = []string{
		"买入", "看好", "抄底", "加仓", "做多", "增持", "推荐",
		"牛市", "涨停", "突破", "创新高", "业绩超预期", "低估",
		"布局", "机会", "价值", "成长", "稳了", "必涨", "翻倍",
	}
	bearishKeywords = []string{
		"卖出", "看空", "逃顶", "减仓", "做空", "减持", "警告",
		"熊市", "跌停", "破位", "创新低", "业绩不及", "高估",
		"止损", "风险", "泡沫", "小心", "别碰", "腰斩", "割肉",
	}
)

// calcSentiment 根据帖子内容计算情绪评分 (-100 ~ +100)
func calcSentiment(title, content string, likes, comments, views int) float64 {
	text := title + " " + content
	score := 0.0

	// 关键词打分
	for _, kw := range bullishKeywords {
		if strings.Contains(text, kw) {
			score += 10
		}
	}
	for _, kw := range bearishKeywords {
		if strings.Contains(text, kw) {
			score -= 10
		}
	}

	// 参与度打分（点赞、评论的加权）
	engagement := float64(likes)*0.5 + float64(comments)*1.0
	if engagement > 0 {
		// 参与度高说明关注度大，情绪更可信
		score = score * (1 + engagement/500.0)
	}

	// 归一化到 -100 ~ +100
	score = max(-100, min(100, score))
	return score
}

// extractKeywords 从文本中提取关键词
func extractKeywords(text string) []string {
	keywords := make([]string, 0)
	textLower := strings.ToLower(text)

	all := append(bullishKeywords, bearishKeywords...)
	for _, kw := range all {
		if strings.Contains(textLower, strings.ToLower(kw)) {
			keywords = append(keywords, kw)
		}
	}

	return uniqStrings(keywords)
}

// uniqStrings 去重
func uniqStrings(ss []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// CleanHTML 去除HTML标签
func CleanHTML(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

// AggregateSentiment 聚合多个帖子的情绪
func AggregateSentiment(posts []Post) float64 {
	if len(posts) == 0 {
		return 0
	}
	var total float64
	for _, p := range posts {
		total += p.Sentiment
	}
	return total / float64(len(posts))
}
