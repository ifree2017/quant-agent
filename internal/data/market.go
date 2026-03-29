package data

// MarketType 市场类型
type MarketType string

const (
	MarketAShare  MarketType = "a_share"  // A股
	MarketFuture  MarketType = "future"  // 期货
	MarketCrypto  MarketType = "crypto"   // 数字货币
)

// BarMeta K线元数据
type BarMeta struct {
	Market   MarketType // 市场类型
	Symbol   string    // 合约代码
	Exchange string    // 交易所代码
	Period   string    // 周期：1m/5m/1h/1d
}
