# QuantAgent — 接口评审记录 v1.1

> 版本：v1.1 | 日期：2026-03-29
> 更新：全 Go 架构，移除 gRPC
> 状态：待三方签字

---

## 1. 接口清单（全 Go Gin HTTP）

### 1.1 REST API（Gin — Go 层）

#### POST `/api/style/analyze`
**描述**：分析用户历史交易记录，输出风格画像

**请求体**：
```json
{
  "records": [
    {
      "date": "2025-01-02",
      "action": "buy",
      "symbol": "000001",
      "price": 10.50,
      "quantity": 100
    }
  ]
}
```

**响应**：
```json
{
  "riskScore": 65.0,
  "style": "积极",
  "tradeFrequency": "中频",
  "avgHoldDays": 15.3,
  "maxDrawdownTolerance": 10.0
}
```

**实现路径**：
```
Handler: internal/api/handler/style.go → Analyze()
  → internal/ai/client.go → LLM API (OpenAI/Groq)
  → Parse JSON → Response
```

---

#### POST `/api/strategy/generate`
**描述**：基于风格画像生成策略规则

**请求体**：
```json
{
  "styleProfile": {
    "riskScore": 65.0,
    "style": "积极",
    "tradeFrequency": "中频",
    "avgHoldDays": 15.3,
    "maxDrawdownTolerance": 10.0
  },
  "market": "A-share",
  "symbol": "000001"
}
```

**响应**：
```json
{
  "strategyId": "uuid-v4",
  "name": "积极型 MA+RSI 策略",
  "rules": {
    "indicators": ["MA", "RSI"],
    "params": {
      "MA_period": 20,
      "RSI_period": 14,
      "RSI_overbought": 70,
      "RSI_oversold": 30
    },
    "entry": { "type": "cross", "condition": "MA_cross_RSI" },
    "exit": { "type": "stop_loss", "value": 0.05 },
    "position": { "type": "fixed", "value": 0.3 }
  },
  "backtestReport": {
    "totalReturn": 0.234,
    "annualReturn": 0.18,
    "sharpeRatio": 1.45,
    "maxDrawdown": 0.082,
    "winRate": 0.58,
    "profitLossRatio": 1.72,
    "totalTrades": 24
  }
}
```

---

#### POST `/api/backtest`
**描述**：对指定策略执行回测

**请求体**：
```json
{
  "strategyId": "uuid-v4",
  "symbol": "000001",
  "days": 60,
  "initialCash": 50000.0
}
```

**响应**：
```json
{
  "jobId": "uuid-v4",
  "status": "running"
}
```

---

#### GET `/api/backtest/{jobId}`
**描述**：查询回测结果

**响应**：
```json
{
  "jobId": "uuid-v4",
  "status": "completed",
  "report": {
    "totalReturn": 0.234,
    "annualReturn": 0.18,
    "sharpeRatio": 1.45,
    "maxDrawdown": 0.082,
    "winRate": 0.58,
    "profitLossRatio": 1.72,
    "totalTrades": 24,
    "trades": [...]
  }
}
```

---

#### GET `/api/strategies`
**描述**：策略列表

**响应**：
```json
{
  "strategies": [
    {
      "id": "uuid-v4",
      "name": "积极型 MA+RSI 策略",
      "style": "积极",
      "version": 1,
      "createdAt": "2025-03-29T12:00:00Z"
    }
  ]
}
```

---

## 2. 评审检查点

| 接口 | 实现路径 | 输入验证 | 输出格式 | 错误处理 | 状态 |
|------|---------|---------|---------|---------|------|
| POST /api/style/analyze | `handler/style.go` | ✅ records 非空 | ✅ JSON | ✅ 400/500 | 待评 |
| POST /api/strategy/generate | `handler/strategy.go` | ✅ styleProfile 必填 | ✅ JSON | ✅ 400/500 | 待评 |
| POST /api/backtest | `handler/backtest.go` | ✅ strategyId + symbol | ✅ JSON | ✅ 400/404 | 待评 |
| GET /api/backtest/{jobId} | `handler/backtest.go` | ✅ jobId 格式 | ✅ JSON | ✅ 404 | 待评 |
| GET /api/strategies | `handler/strategy.go` | — | ✅ JSON | ✅ 500 | 待评 |
| DELETE /api/strategies/:id | `handler/strategy.go` | ✅ id 存在 | ✅ JSON | ✅ 404 | 待评 |

---

## 3. 评审签字

| 角色 | 姓名 | 签字 | 日期 | 意见 |
|------|------|------|------|------|
| 前端 RD | | | | |
| 后端 RD | | | | |
| PM | | | | |

---

## 4. 接口变更记录

| 日期 | 变更内容 | 变更人 | 签字 |
|------|---------|--------|------|
| 2026-03-29 | 初始版本（全 Go） | RD Lead | |
| 2026-03-29 | 更新：移除 gRPC，全 Go Gin HTTP | RD Lead | |
