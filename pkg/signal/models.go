package signal

import (
	"time"
)

// SignalType 交易信号类型
type SignalType string

const (
	SignalBuy   SignalType = "BUY"
	SignalSell  SignalType = "SELL"
	SignalHold  SignalType = "HOLD"
	SignalWatch SignalType = "WATCH" // 关注（情绪异动）
)

// Signal 交易信号
type Signal struct {
	StockCode  string     // 股票代码 "000001"
	SignalType SignalType // BUY | SELL | HOLD | WATCH
	Confidence float64    // 置信度 0-1
	Source     string     // 信号来源 "social_sentiment" | "sector_rotation" | "news_alert"
	Score      float64    // 综合情绪分 -100~+100
	Weight     float64    // 信号权重
	Message    string     // 信号描述
	CreatedAt  time.Time  // 创建时间
}

// SectorSignal 板块轮动信号
type SectorSignal struct {
	SectorName  string    // 板块名称 "新能源" | "半导体" | ...
	HeatScore   float64   // 热力值 0-100
	Direction   string    // 轮动方向 "emerging" | "fading" | "stable"
	StockCodes  []string  // 相关股票代码
	Keywords    []string  // 触发关键词
	ChangePct   float64   // 涨跌幅 %
	Source      string    // 信号来源
	CreatedAt   time.Time // 创建时间
}

// AlertLevel 预警等级
type AlertLevel string

const (
	AlertLow      AlertLevel = "LOW"
	AlertMedium   AlertLevel = "MEDIUM"
	AlertHigh     AlertLevel = "HIGH"
	AlertCritical AlertLevel = "CRITICAL"
)

// AlertSignal 舆情预警信号
type AlertSignal struct {
	StockCode   string      // 股票代码
	AlertLevel  AlertLevel  // 预警等级 LOW/MEDIUM/HIGH/CRITICAL
	Title       string      // 预警标题
	Description string      // 预警描述
	Keywords    []string    // 触发关键词
	NewsCount   int         // 相关新闻数量
	Duration    string      // 持续时间 "短期" | "中期" | "长期"
	Source      string      // 信号来源
	CreatedAt   time.Time   // 创建时间
}

// PlatformWeight 平台权重配置
var PlatformWeight = map[string]float64{
	"eastmoney": 0.30, // 东财 30%
	"xueqiu":    0.25, // 雪球 25%
	"wechat":    0.20, // 微信 20%
	"douyin":    0.15, // 抖音 15%
	"news":      0.10, // 新闻 10%
}

// PlatformNames 平台名称列表
var PlatformNames = []string{"eastmoney", "xueqiu", "wechat", "douyin", "news"}

// NegativeKeywords 负面关键词（用于舆情预警）
var NegativeKeywords = []string{
	"黑天鹅", "减持", "解禁", "业绩下滑", "监管", "调查", "处罚",
	"退市", "亏损", "债务", "违约", "造假", "欺诈", "爆雷",
	"踩雷", "大幅下跌", "暴跌", "破发", "套牢", "割肉", "减持",
}

// SectorKeywords 板块关键词映射
var SectorKeywords = map[string][]string{
	"新能源":     {"新能源", "锂电", "光伏", "储能", "电动汽车", "电动车", "电池", "比亚迪", "宁德时代", "特斯拉"},
	"半导体":     {"半导体", "芯片", "集成电路", "晶圆", "光刻", "封测", "设备", "材料", "AI芯片"},
	"人工智能":   {"人工智能", "AI", "大模型", "机器学习", "深度学习", "ChatGPT", "LLM", "算力"},
	"医药":      {"医药", "医疗器械", "创新药", "疫苗", "中药", "生物药", "CRO", "CXO"},
	"消费":      {"消费", "食品饮料", "白酒", "家电", "汽车", "零售", "旅游", "酒店"},
	"金融":      {"金融", "银行", "保险", "证券", "券商", "信托", "基金"},
	"房地产":     {"房地产", "地产", "万科", "保利", "建筑", "建材", "家居"},
	"军工":      {"军工", "国防", "航天", "航空", "船舶", "无人机", "导弹"},
	"元宇宙":     {"元宇宙", "VR", "AR", "虚拟现实", "游戏", "NFT", "Web3"},
	"数字经济":    {"数字经济", "云计算", "数据中心", "网络安全", "软件", "操作系统", "信创"},
}

// ScoreThresholds 分数阈值配置
var ScoreThresholds = struct {
	BuyScore     float64
	BuyConf      float64
	SellScore    float64
	SellConf     float64
	WatchScore   float64
	WatchConf    float64
}{
	BuyScore:   60,
	BuyConf:    0.7,
	SellScore:  -60,
	SellConf:   0.7,
	WatchScore: 30,
	WatchConf:  0.5,
}
