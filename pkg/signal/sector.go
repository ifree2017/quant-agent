package signal

import (
	"context"
	"strings"
	"sync"
	"time"

	"quant-agent/pkg/social"
)

// DetectSectorHot 检测板块轮动热点
// 逻辑：
//  1. 获取抖音热榜 + 东财热点
//  2. 关键词匹配板块：["新能源","锂电","光伏"]→ 新能源板块
//  3. 热度排名变化 → 板块轮动信号
//  4. 输出：板块名、热力值、轮动方向(emerging/fading)
func (e *Engine) DetectSectorHot() ([]SectorSignal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return e.DetectSectorHotWithContext(ctx)
}

// DetectSectorHotWithContext 带超时的板块检测
func (e *Engine) DetectSectorHotWithContext(ctx context.Context) ([]SectorSignal, error) {
	type hotResult struct {
		platform string
		items    interface{}
		err      error
	}

	resultCh := make(chan hotResult, 2)
	var wg sync.WaitGroup

	// 并发获取抖音和东财热点
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			resultCh <- hotResult{err: ctx.Err()}
			return
		default:
		}

		items, err := social.FetchDouyinHot()
		resultCh <- hotResult{platform: "douyin", items: items, err: err}
	}()

	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			resultCh <- hotResult{err: ctx.Err()}
			return
		default:
		}

		client := social.NewEastmoneyClient()
		items, err := client.FetchMarketHot()
		resultCh <- hotResult{platform: "eastmoney", items: items, err: err}
	}()

	// 等待完成
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 收集并聚合结果
	sectorMap := make(map[string]*SectorSignal)
	keywordsSeen := make(map[string]map[string]bool) // sector -> keywords

	for result := range resultCh {
		if result.err != nil {
			continue
		}

		switch result.platform {
		case "douyin":
			if items, ok := result.items.([]social.DouyinHotItem); ok {
				e.processDouyinHot(items, sectorMap, keywordsSeen)
			}
		case "eastmoney":
			if items, ok := result.items.([]social.MarketHotItem); ok {
				e.processEastmoneyHot(items, sectorMap, keywordsSeen)
			}
		}
	}

	// 转换为结果数组
	signals := make([]SectorSignal, 0, len(sectorMap))
	for name, signal := range sectorMap {
		signal.SectorName = name
		signal.Source = "sector_rotation"
		signal.CreatedAt = time.Now()

		// 设置轮动方向
		signal.Direction = e.detectDirection(signal)

		// 提取关键词
		if kwMap, ok := keywordsSeen[name]; ok {
			for kw := range kwMap {
				signal.Keywords = append(signal.Keywords, kw)
			}
		}

		signals = append(signals, *signal)
	}

	// 按热力值排序
	sortSectorSignals(signals)

	return signals, nil
}

// processDouyinHot 处理抖音热榜数据
func (e *Engine) processDouyinHot(items []social.DouyinHotItem, sectorMap map[string]*SectorSignal, keywordsSeen map[string]map[string]bool) {
	for _, item := range items {
		// 跳过广告
		if item.IsAd {
			continue
		}

		// 关键词匹配板块
		sectors := e.matchSectors(item.Title)
		for _, sector := range sectors {
			if _, ok := sectorMap[sector]; !ok {
				sectorMap[sector] = &SectorSignal{}
				keywordsSeen[sector] = make(map[string]bool)
			}

			s := sectorMap[sector]
			s.StockCodes = append(s.StockCodes, item.Title)
			s.HeatScore += float64(item.HotValue) / 1e6 // 归一化

			// 提取关键词
			for _, kw := range e.extractHotKeywords(item.Title) {
				keywordsSeen[sector][kw] = true
			}
		}
	}
}

// processEastmoneyHot 处理东财热点数据
func (e *Engine) processEastmoneyHot(items []social.MarketHotItem, sectorMap map[string]*SectorSignal, keywordsSeen map[string]map[string]bool) {
	for _, item := range items {
		// 关键词匹配板块
		searchText := item.StockName + " " + item.StockCode
		sectors := e.matchSectors(searchText)

		for _, sector := range sectors {
			if _, ok := sectorMap[sector]; !ok {
				sectorMap[sector] = &SectorSignal{}
				keywordsSeen[sector] = make(map[string]bool)
			}

			s := sectorMap[sector]
			s.StockCodes = append(s.StockCodes, item.StockCode)
			s.ChangePct += item.ChangePct
			s.HeatScore += item.HeatScore

			// 提取关键词
			keywordsSeen[sector][item.StockName] = true
		}
	}
}

// matchSectors 关键词匹配板块
func (e *Engine) matchSectors(text string) []string {
	text = strings.ToLower(text)
	matched := make([]string, 0)

	for sector, keywords := range SectorKeywords {
		for _, kw := range keywords {
			if strings.Contains(text, strings.ToLower(kw)) {
				matched = append(matched, sector)
				break
			}
		}
	}

	// 如果没有匹配，返回空数组
	return matched
}

// extractHotKeywords 从热门话题中提取关键词
func (e *Engine) extractHotKeywords(text string) []string {
	keywords := make([]string, 0)

	// 简单分词：按常见分隔符分割
	separators := []string{" ", "｜", "|", "#", "【", "】", "（", "）", "(", ")"}
	processed := text
	for _, sep := range separators {
		parts := strings.Split(processed, sep)
		if len(parts) > 1 {
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if len(p) >= 2 && len(p) <= 10 {
					keywords = append(keywords, p)
				}
			}
		}
	}

	// 如果没有提取到，返回原文本作为关键词
	if len(keywords) == 0 {
		text = strings.TrimSpace(text)
		if len(text) >= 2 && len(text) <= 20 {
			keywords = append(keywords, text)
		}
	}

	return keywords
}

// detectDirection 检测轮动方向
func (e *Engine) detectDirection(signal *SectorSignal) string {
	// 基于涨跌幅判断
	if signal.ChangePct > 5 {
		return "emerging" // 新兴热点
	}
	if signal.ChangePct < -3 {
		return "fading" // 衰退热点
	}
	return "stable" // 稳定
}

// sortSectorSignals 按热力值排序
func sortSectorSignals(signals []SectorSignal) {
	for i := 0; i < len(signals)-1; i++ {
		for j := i + 1; j < len(signals); j++ {
			if signals[j].HeatScore > signals[i].HeatScore {
				signals[i], signals[j] = signals[j], signals[i]
			}
		}
	}
}
