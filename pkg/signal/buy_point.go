package signal

import (
	"fmt"
	"time"
)

// BuyPointPatterns 买点模式
type BuyPointPatterns struct {
	Pullback bool `json:"pullback"` // 回踩模式
	Breakout bool `json:"breakout"` // 突破模式
	SecondWave bool `json:"second_wave"` // 二波模式
	Flow     bool `json:"flow"`     // 主动买入比>1.2
	Pressure bool `json:"pressure"` // 无卖盘压制
}

// BuyPointSignal 买点信号
type BuyPointSignal struct {
	Code      string          `json:"code"`
	Signal    BuyPointLevel   `json:"signal"`
	Score     int             `json:"score"`
	Patterns  BuyPointPatterns `json:"patterns"`
	AntiChase bool            `json:"anti_chase"`
	Action    TradeAction     `json:"action"`
	Timestamp time.Time       `json:"timestamp"`
}

// BuyPointLevel 买点级别
type BuyPointLevel string

const (
	BuyPointStrong BuyPointLevel = "STRONG_BUY"
	BuyPointNormal BuyPointLevel = "BUY"
	BuyPointNone   BuyPointLevel = "NO_BUY"
)

// BuyPointConfig 买点配置
type BuyPointConfig struct {
	PullbackMinPct   float64 // 回踩最小幅度（默认2%）
	PullbackMaxPct   float64 // 回踩最大幅度（默认5%）
	BreakoutVolMult  float64 // 突破放量倍数（默认1.5）
	SecondWaveThresh float64 // 二波阈值（默认0.95，即95%）
	FlowRatio        float64 // 主动买入比阈值（默认1.2）
	PressureRatio    float64 // 无压制比率（默认1.2）
}

// DefaultBuyPointConfig 默认买点配置
var DefaultBuyPointConfig = BuyPointConfig{
	PullbackMinPct:   2.0,
	PullbackMaxPct:   5.0,
	BreakoutVolMult:  1.5,
	SecondWaveThresh: 0.95,
	FlowRatio:        1.2,
	PressureRatio:    1.2,
}

// CalculateBuyPointScore 计算买点评分
func CalculateBuyPointScore(patterns BuyPointPatterns, config *BuyPointConfig) (int, BuyPointLevel, TradeAction) {
	if config == nil {
		config = &DefaultBuyPointConfig
	}

	score := 0

	// pullback: 回踩幅度2-5%，不破均线/VWAP，权重 +2
	if patterns.Pullback {
		score += 2
	}

	// breakout: 突破前高+放量（>1.5倍均量），权重 +3
	if patterns.Breakout {
		score += 3
	}

	// second_wave: 接近前高（>95%）+量能再放大，权重 +3
	if patterns.SecondWave {
		score += 3
	}

	// flow: 主动买入比>1.2，权重 +2
	if patterns.Flow {
		score += 2
	}

	// pressure: 无卖盘压制（ask<bid*1.2），权重 +1
	if patterns.Pressure {
		score += 1
	}

	// 评分逻辑
	var level BuyPointLevel
	var action TradeAction

	if score >= 6 {
		level = BuyPointStrong
		action = ActionBuyAll
	} else if score >= 4 {
		level = BuyPointNormal
		action = ActionAddPos
	} else {
		level = BuyPointNone
		action = ActionWait
	}

	return score, level, action
}

// AntiChase 防追涨检查：连续大阳线（>2%）禁止买入
func AntiChase(bars []interface{}) bool {
	if len(bars) < 2 {
		return false
	}

	// 支持两种格式：[]Bar 或 []interface{}
	lastBar := getBar(bars[len(bars)-1])
	prevBar := getBar(bars[len(bars)-2])

	if lastBar == nil || prevBar == nil {
		return false
	}

	// 连续大阳线禁止买入
	return lastBar.Close > prevBar.Close*1.02
}

// getBar 尝试从多种格式获取Bar数据
func getBar(v interface{}) *BarPoint {
	switch b := v.(type) {
	case BarPoint:
		return &b
	case *BarPoint:
		return b
	case map[string]interface{}:
		bar := &BarPoint{}
		if o, ok := b["open"].(float64); ok {
			bar.Open = o
		}
		if h, ok := b["high"].(float64); ok {
			bar.High = h
		}
		if l, ok := b["low"].(float64); ok {
			bar.Low = l
		}
		if c, ok := b["close"].(float64); ok {
			bar.Close = c
		}
		if vol, ok := b["volume"].(float64); ok {
			bar.Volume = int64(vol)
		}
		return bar
	}
	return nil
}

// BarPoint 简化Bar结构用于计算
type BarPoint struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// CheckPullback 检查回踩模式
func CheckPullback(currentPrice, maPrice float64) bool {
	if maPrice <= 0 {
		return false
	}
	pct := (currentPrice - maPrice) / maPrice * 100
	return pct >= DefaultBuyPointConfig.PullbackMinPct && pct <= DefaultBuyPointConfig.PullbackMaxPct
}

// CheckBreakout 检查突破模式（前高+放量）
func CheckBreakout(currentHigh, prevHigh, currentVol, avgVol float64) bool {
	if avgVol <= 0 {
		return false
	}
	// 突破前高
	highBreak := currentHigh > prevHigh
	// 放量 > 1.5倍均量
	volBreak := currentVol >= avgVol*DefaultBuyPointConfig.BreakoutVolMult
	return highBreak && volBreak
}

// CheckSecondWave 检查二波模式（接近前高+量能再放大）
func CheckSecondWave(currentPrice, prevHigh, currentVol, prevVol float64) bool {
	if prevHigh <= 0 {
		return false
	}
	// 接近前高 > 95%
	nearHigh := currentPrice >= prevHigh*DefaultBuyPointConfig.SecondWaveThresh
	// 量能再放大
	volExpand := currentVol > prevVol
	return nearHigh && volExpand
}

// CheckFlow 检查主动买入比
func CheckFlow(buyVolume, sellVolume float64) bool {
	if sellVolume <= 0 {
		return buyVolume > 0
	}
	return (buyVolume / sellVolume) > DefaultBuyPointConfig.FlowRatio
}

// CheckPressure 检查无卖盘压制
func CheckPressure(askVolume, bidVolume float64) bool {
	if bidVolume <= 0 {
		return true
	}
	return askVolume < bidVolume*DefaultBuyPointConfig.PressureRatio
}

// NewBuyPointSignal 创建买点信号
func NewBuyPointSignal(code string, patterns BuyPointPatterns, antiChase bool) *BuyPointSignal {
	score, signal, action := CalculateBuyPointScore(patterns, nil)
	return &BuyPointSignal{
		Code:      code,
		Signal:    signal,
		Score:     score,
		Patterns:  patterns,
		AntiChase: antiChase,
		Action:    action,
		Timestamp: time.Now(),
	}
}

// String 返回信号描述
func (s *BuyPointSignal) String() string {
	return fmt.Sprintf("[%s] score=%d patterns={pullback:%v breakout:%v second_wave:%v flow:%v pressure:%v} anti_chase:%v action:%s",
		s.Signal, s.Score, s.Patterns.Pullback, s.Patterns.Breakout, s.Patterns.SecondWave,
		s.Patterns.Flow, s.Patterns.Pressure, s.AntiChase, s.Action)
}
