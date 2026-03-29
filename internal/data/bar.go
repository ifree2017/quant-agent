package data

import "time"

// Bar K线数据
type Bar struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// MAData 均线数据
type MAData struct {
	MA5   float64
	MA10  float64
	MA20  float64
	MA60  float64
}

// RSIData RSI数据
type RSIData struct {
	RSI6  float64
	RSI14 float64
}

// MACDData MACD数据
type MACDData struct {
	MACD    float64
	Signal  float64
	Histogram float64
}

// BollingerData 布林带数据
type BollingerData struct {
	Upper  float64
	Middle float64
	Lower  float64
}

// IndicatorStore 指标存储（用于跨bar计算）
type IndicatorStore struct {
	Bars        []Bar
	MA          MAData
	RSI         RSIData
	MACD        MACDData
	Bollinger   BollingerData
	ClosePrices []float64 // 用于计算RSI
}
