package signal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"quant-agent/pkg/social"
)

// Engine 信号引擎：聚合多平台情绪 → 输出交易信号
type Engine struct {
	socialClient *social.Client // 注入现有social包（用于共享HTTPClient和RateLimiter）
}

// NewEngine 创建信号引擎
func NewEngine() *Engine {
	return &Engine{
		socialClient: social.NewClient(),
	}
}

// GenerateSignals 为指定股票生成交易信号
// 逻辑：
//  1. 并发拉取各平台情绪数据
//  2. 加权汇总情绪分（东财30% + 雪球25% + 微信20% + 抖音15% + 新闻10%）
//  3. 应用信号规则
//  4. 提取关键词作为信号描述
func (e *Engine) GenerateSignals(stockCode string) ([]Signal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return e.GenerateSignalsWithContext(ctx, stockCode)
}

// GenerateSignalsWithContext 带超时的信号生成
func (e *Engine) GenerateSignalsWithContext(ctx context.Context, stockCode string) ([]Signal, error) {
	// 并发拉取各平台数据
	type platformResult struct {
		platform string
		data     *social.SentimentData
		err      error
	}

	resultCh := make(chan platformResult, len(PlatformNames))
	var wg sync.WaitGroup

	// 并发获取各平台情绪数据
	for _, platform := range PlatformNames {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				resultCh <- platformResult{platform: p, err: ctx.Err()}
				return
			default:
			}

			var data *social.SentimentData
			var err error

			switch p {
			case "eastmoney":
				client := social.NewEastmoneyClient()
				data, err = client.FetchStockSentiment(stockCode)
			case "xueqiu":
				client := social.NewXueqiuClient()
				data, err = client.FetchStockSentiment(stockCode)
			case "wechat":
				// 微信没有直接的个股情绪API，跳过
				resultCh <- platformResult{platform: p, data: nil, err: nil}
				return
			case "douyin":
				// 抖音没有直接的个股情绪API，跳过
				resultCh <- platformResult{platform: p, data: nil, err: nil}
				return
			case "news":
				// 新闻也没有直接的个股情绪API，跳过
				resultCh <- platformResult{platform: p, data: nil, err: nil}
				return
			}

			resultCh <- platformResult{platform: p, data: data, err: err}
		}(platform)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 收集结果
	var weightedScore float64
	var totalWeight float64
	var allKeywords []string
	var maxConfidence float64

	for result := range resultCh {
		if result.err != nil || result.data == nil {
			continue
		}

		weight := PlatformWeight[result.platform]
		weightedScore += result.data.SentimentScore * weight
		totalWeight += weight

		allKeywords = append(allKeywords, result.data.Keywords...)

		// 计算置信度（基于帖子数量和热度）
		conf := e.calcConfidence(result.data)
		if conf > maxConfidence {
			maxConfidence = conf
		}
	}

	// 归一化
	if totalWeight > 0 {
		weightedScore = weightedScore / totalWeight * (1 / totalWeight)
	}

	// 限制范围
	score := max(-100.0, min(100.0, weightedScore))

	// 生成信号
	signalType, conf := e.applyRules(score, maxConfidence)
	message := e.generateMessage(score, allKeywords, stockCode)

	signal := Signal{
		StockCode:  stockCode,
		SignalType: signalType,
		Confidence: conf,
		Source:     "social_sentiment",
		Score:      score,
		Weight:     1.0,
		Message:    message,
		CreatedAt:  time.Now(),
	}

	return []Signal{signal}, nil
}

// calcConfidence 计算置信度
func (e *Engine) calcConfidence(data *social.SentimentData) float64 {
	if data == nil {
		return 0
	}

	// 基于帖子数量和热度计算置信度
	postFactor := float64(data.PostCount) / float64(maxInt(data.PostCount, 20))
	heatFactor := data.HeatScore / 100.0

	conf := (postFactor*0.3 + heatFactor*0.7)
	return max(0, min(1, conf))
}

// applyRules 应用信号规则
func (e *Engine) applyRules(score, confidence float64) (SignalType, float64) {
	if score > ScoreThresholds.BuyScore && confidence > ScoreThresholds.BuyConf {
		return SignalBuy, confidence
	}
	if score < ScoreThresholds.SellScore && confidence > ScoreThresholds.SellConf {
		return SignalSell, confidence
	}
	if score > ScoreThresholds.WatchScore && confidence > ScoreThresholds.WatchConf {
		return SignalWatch, confidence
	}
	return SignalHold, confidence
}

// generateMessage 生成信号描述
func (e *Engine) generateMessage(score float64, keywords []string, stockCode string) string {
	var sb strings.Builder

	// 情绪方向
	var direction string
	if score > 60 {
		direction = "强烈看多"
	} else if score > 30 {
		direction = "偏多"
	} else if score > 0 {
		direction = "略偏多"
	} else if score > -30 {
		direction = "略偏空"
	} else if score > -60 {
		direction = "偏空"
	} else {
		direction = "强烈看空"
	}

	sb.WriteString(fmt.Sprintf("[%s] 综合情绪分: %.1f", direction, score))

	// 添加关键词
	if len(keywords) > 0 {
		uniqueKeywords := unique(keywords)
		if len(uniqueKeywords) > 5 {
			uniqueKeywords = uniqueKeywords[:5]
		}
		sb.WriteString(fmt.Sprintf(" | 关键词: %s", strings.Join(uniqueKeywords, ", ")))
	}

	return sb.String()
}

// unique 去重
func unique(ss []string) []string {
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

// max 取最大值
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// min 取最小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxInt 取最大值
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
