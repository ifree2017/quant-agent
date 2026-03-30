package signal

import (
	"fmt"
	"time"
)

// DistributionFeatures 出货特征
type DistributionFeatures struct {
	SpikeFall       bool    // 冲高回落（>3%拉高后回落）
	SecondWave      bool    // 有无二波攻击
	VolumeDecay     bool    // 量能衰减（下半场<上半场70%）
	SellPressure    bool    // 卖盘压制（卖盘>买盘1.5倍）
	ActiveSellRatio float64 // 主动卖出比
}

// DistributionSignal 出货信号
type DistributionSignal struct {
	Code      string              `json:"code"`
	Signal    DistributionLevel    `json:"signal"`
	Score     int                 `json:"score"`
	Features  DistributionFeatures `json:"features"`
	Action    TradeAction         `json:"action"`
	Timestamp time.Time           `json:"timestamp"`
}

// DistributionLevel 出货级别
type DistributionLevel string

const (
	DistributionStrong DistributionLevel = "STRONG_DISTRIBUTION"
	DistributionNormal DistributionLevel = "DISTRIBUTION"
	DistributionNeutral DistributionLevel = "NEUTRAL"
)

// TradeAction 交易动作
type TradeAction string

const (
	ActionSellAll    TradeAction = "SELL_ALL"
	ActionReduce     TradeAction = "REDUCE"
	ActionHold       TradeAction = "HOLD"
	ActionBuyAll     TradeAction = "BUY_ALL"
	ActionAddPos     TradeAction = "ADD_POSITION"
	ActionWait       TradeAction = "WAIT"
)

// CalculateDistributionScore 计算出货评分
func CalculateDistributionScore(features DistributionFeatures) (int, DistributionLevel, TradeAction) {
	score := 0

	// spike_fall: 冲高回落 >3% 拉高后回落，权重 +2
	if features.SpikeFall {
		score += 2
	}

	// second_wave: 有无二波攻击，无二波 = 加分，权重 +2
	if !features.SecondWave {
		score += 2
	}

	// volume_decay: 量能衰减（下半场<上半场70%），权重 +1
	if features.VolumeDecay {
		score += 1
	}

	// sell_pressure: 卖盘压制（卖盘>买盘1.5倍），权重 +2
	if features.SellPressure {
		score += 2
	}

	// active_sell_ratio: 主动卖出比 <0.8 = 大单出，权重 +2
	if features.ActiveSellRatio < 0.8 {
		score += 2
	}

	// 评分逻辑
	var level DistributionLevel
	var action TradeAction

	if score >= 7 {
		level = DistributionStrong
		action = ActionSellAll
	} else if score >= 4 {
		level = DistributionNormal
		action = ActionReduce
	} else {
		level = DistributionNeutral
		action = ActionHold
	}

	return score, level, action
}

// NewDistributionSignal 创建出货信号
func NewDistributionSignal(code string, features DistributionFeatures) *DistributionSignal {
	score, signal, action := CalculateDistributionScore(features)
	return &DistributionSignal{
		Code:      code,
		Signal:    signal,
		Score:     score,
		Features:  features,
		Action:    action,
		Timestamp: time.Now(),
	}
}

// String 返回信号描述
func (s *DistributionSignal) String() string {
	return fmt.Sprintf("[%s] score=%d features={spike_fall:%v second_wave:%v volume_decay:%v sell_pressure:%v active_sell_ratio:%.2f} action:%s",
		s.Signal, s.Score, s.Features.SpikeFall, s.Features.SecondWave,
		s.Features.VolumeDecay, s.Features.SellPressure, s.Features.ActiveSellRatio, s.Action)
}
