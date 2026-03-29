package strategy

import (
	"context"
	"time"

	"quant-agent/internal/data"
	"quant-agent/internal/model"
	"quant-agent/pkg/signal"
)

// SocialMomentumStrategy 基于社交情绪的预置策略
// 规则：
//   - 雪球/东财情绪分 > 60 → 买入
//   - 持仓期间情绪分 < 30 → 止盈
//   - 新闻负面预警 → 止损
// 回测表现：适用于短线情绪驱动型股票
type SocialMomentumStrategy struct {
	Executor
	signalEngine *signal.Engine
	position     float64 // 持仓状态：0=空仓, >0=持仓
	entryScore   float64 // 入场时的情绪分
}

// NewSocialMomentumStrategy 创建社交情绪动量策略
func NewSocialMomentumStrategy(rules model.StrategyRules) *SocialMomentumStrategy {
	return &SocialMomentumStrategy{
		Executor:     *NewExecutor(rules),
		signalEngine: signal.NewEngine(),
		position:     0,
		entryScore:   0,
	}
}

// SocialStrategyRules 社交策略预置规则
var SocialStrategyRules = model.StrategyRules{
	Indicators: []string{"MA", "RSI"},
	Params: map[string]float64{
		"MA_period":        20,
		"RSI_period":       14,
		"RSI_overbought":   70,
		"RSI_oversold":     30,
		"sentiment_buy":    60,   // 买入情绪阈值
		"sentiment_sell":   30,   // 止盈情绪阈值
		"sentiment_stop":   -60,  // 止损情绪阈值
		"confidence_thres": 0.7, // 置信度阈值
	},
	Entry: model.EntryRule{
		Type:      "social_sentiment",
		Condition: "sentiment_buy",
	},
	Exit: model.ExitRule{
		Type:  "social_sentiment",
		Value: 30,
	},
	Position: model.PositionRule{
		Type:  "fixed",
		Value: 0.2,
	},
}

// UpdateWithSocial 更新策略（包含社交情绪信号）
func (s *SocialMomentumStrategy) UpdateWithSocial(bar data.Bar, stockCode string) {
	// 先更新基础指标
	s.Update(bar)

	// 获取社交情绪信号
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	signals, err := s.signalEngine.GenerateSignalsWithContext(ctx, stockCode)
	if err != nil || len(signals) == 0 {
		return
	}

	socialSignal := signals[0]

	// 更新持仓状态
	if s.position == 0 && socialSignal.SignalType == signal.SignalBuy {
		// 空仓且买入信号 → 开仓
		s.position = 1
		s.entryScore = socialSignal.Score
	} else if s.position > 0 {
		// 持仓中，检查止盈/止损

		// 止损：负面预警
		alerts, _ := s.signalEngine.DetectAlertsWithContext(ctx, stockCode)
		if len(alerts) > 0 && (alerts[0].AlertLevel == signal.AlertCritical || alerts[0].AlertLevel == signal.AlertHigh) {
			s.position = 0 // 止损
			s.entryScore = 0
			return
		}

		// 止盈：情绪分低于阈值
		if socialSignal.Score < s.rules.Params["sentiment_sell"] {
			s.position = 0 // 止盈
			s.entryScore = 0
			return
		}
	}
}

// SignalWithSocial 带社交情绪的交易信号
func (s *SocialMomentumStrategy) SignalWithSocial(stockCode string) string {
	// 获取社交情绪信号
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	signals, err := s.signalEngine.GenerateSignalsWithContext(ctx, stockCode)
	if err != nil || len(signals) == 0 {
		return s.Signal() // 回退到基础信号
	}

	socialSignal := signals[0]
	confidence := s.rules.Params["confidence_thres"]

	// 高置信度信号优先
	if socialSignal.Confidence >= confidence {
		switch socialSignal.SignalType {
		case signal.SignalBuy:
			return "buy"
		case signal.SignalSell:
			return "sell"
		case signal.SignalWatch:
			// WATCH信号：如果当前持仓则保持，不开新仓
			if s.position > 0 {
				return "hold"
			}
			return "hold"
		}
	}

	// 回退到基础技术信号
	return s.Signal()
}

// GetPosition 获取当前持仓状态
func (s *SocialMomentumStrategy) GetPosition() float64 {
	return s.position
}

// GetEntryScore 获取入场情绪分
func (s *SocialMomentumStrategy) GetEntryScore() float64 {
	return s.entryScore
}

// Reset 重置策略状态
func (s *SocialMomentumStrategy) Reset() {
	s.position = 0
	s.entryScore = 0
	s.Executor = *NewExecutor(s.rules)
}
