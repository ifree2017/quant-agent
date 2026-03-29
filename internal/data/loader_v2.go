package data

import (
	"os"
	"path/filepath"
	"strings"
)

// LoaderV2 统一数据加载器
type LoaderV2 struct {
	dataDir string
}

// NewLoaderV2 创建V2加载器
func NewLoaderV2(dataDir string) *LoaderV2 {
	return &LoaderV2{dataDir: dataDir}
}

// LoadBarsAdvanced 高级加载（自动识别市场类型）
func (l *LoaderV2) LoadBarsAdvanced(symbol string, days int) ([]Bar, MarketType, error) {
	// 1. 识别市场类型
	market := l.detectMarket(symbol)

	// 2. 查找数据文件
	filename := l.findFile(symbol, market)
	if filename == "" {
		return nil, "", ErrFileNotFound
	}

	// 3. 按市场类型加载
	switch market {
	case MarketAShare:
		return l.loadAShare(filename, days)
	case MarketFuture:
		return l.loadFuture(filename, days)
	case MarketCrypto:
		return l.loadCrypto(filename, days)
	default:
		return nil, "", ErrUnknownMarket
	}
}

// detectMarket 识别市场类型
func (l *LoaderV2) detectMarket(symbol string) MarketType {
	// A股：6位数字
	if len(symbol) == 6 && isDigit(symbol) {
		return MarketAShare
	}
	// 期货：品种代码如 rb2101, HC2105, i., j.
	if strings.HasPrefix(symbol, "rb") ||
		strings.HasPrefix(symbol, "HC") ||
		strings.HasPrefix(symbol, "i.") ||
		strings.HasPrefix(symbol, "j.") {
		return MarketFuture
	}
	// 数字货币：BTC, ETH, BNB, USDT...
	upper := strings.ToUpper(symbol)
	if upper == "BTC" || upper == "ETH" || upper == "BNB" || upper == "USDT" {
		return MarketCrypto
	}
	return MarketAShare // 默认
}

// findFile 查找数据文件
func (l *LoaderV2) findFile(symbol string, market MarketType) string {
	// 优先精确匹配
	patterns := []string{
		filepath.Join(l.dataDir, symbol+".csv"),
		filepath.Join(l.dataDir, strings.ToLower(symbol)+".csv"),
	}
	switch market {
	case MarketAShare:
		patterns = append(patterns, filepath.Join(l.dataDir, "a_share", symbol+".csv"))
	case MarketFuture:
		patterns = append(patterns, filepath.Join(l.dataDir, "future", symbol+".csv"))
	case MarketCrypto:
		patterns = append(patterns, filepath.Join(l.dataDir, "crypto", symbol+".csv"))
	}
	for _, p := range patterns {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// isDigit 检查字符串是否全为数字
func isDigit(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
