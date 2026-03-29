# QuantAgent 智能量化策略系统 — 技术方案 v1.1

> 版本：v1.1 | 日期：2026-03-29
> 更新：全 Go 架构，移除 Python/FastAPI
> 状态：待评审

---

## 1. 系统架构

### 1.1 整体架构（全 Go）

```
┌─────────────────────────────────────────────────────┐
│                   quant-agent（全 Go 单进程）              │
│                                                     │
│  ┌──────────┐    ┌──────────────────────────────┐  │
│  │  Web UI  │───→│       Go API (Gin)          │  │
│  │  (React) │    │  ┌────────┐  ┌───────────┐  │  │
│  └──────────┘    │  │AI Engine│  │ Backtest   │  │  │
│                    │  │(HTTP→LLM)│ │  Engine    │  │  │
│                    │  └────────┘  └───────────┘  │  │
│                    │  ┌────────┐  ┌───────────┐  │  │
│                    │  │Strategy │  │Data Loader │  │  │
│                    │  │ Manager │  │(CSV/SQLite)│  │  │
│                    │  └────────┘  └───────────┘  │  │
│                    └──────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### 1.2 技术栈

| 模块 | 语言 | 框架 | 说明 |
|------|------|------|------|
| API 层 | Go | Gin | HTTP 接口 |
| AI 引擎 | Go | 原生 HTTP Client | 调用 OpenAI/Groq API |
| 回测引擎 | Go | 原生 | 高性能撮合 + 绩效计算 |
| 数据层 | Go | SQLite | 策略存储 + 回测结果 |
| 前端 | TypeScript | React | 策略管理界面 |

### 1.3 模块职责（全 Go）

```
quant-agent（单一 Go 二进制）
├── cmd/
│   └── server/
│       └── main.go        — 程序入口，Gin 启动
├── internal/
│   ├── api/
│   │   ├── server.go     — Gin HTTP 服务器
│   │   ├── handler/      — HTTP Handler（风格/策略/回测）
│   │   └── middleware.go  — 中间件（日志/认证）
│   ├── ai/
│   │   ├── client.go     — LLM HTTP 客户端
│   │   ├── style.go      — 风格分析调用
│   │   └── strategy.go    — 策略生成调用
│   ├── backtest/
│   │   ├── engine.go     — 回测引擎核心
│   │   ├── account.go    — 模拟账户
│   │   ├── portfolio.go  — 持仓管理
│   │   └── config.go     — 回测配置
│   ├── strategy/
│   │   ├── executor.go    — 策略执行器
│   │   ├── ma.go        — MA 指标
│   │   ├── rsi.go       — RSI 指标
│   │   ├── macd.go      — MACD 指标
│   │   └── bollinger.go  — 布林带指标
│   ├── data/
│   │   ├── bar.go       — K线数据结构
│   │   └── loader.go    — CSV 数据加载
│   ├── report/
│   │   ├── metrics.go    — 绩效指标计算
│   │   └── store.go      — SQLite 持久化
│   └── model/
│       ├── style.go      — 风格画像模型
│       ├── strategy.go   — 策略模型
│       └── backtest.go   — 回测结果模型
├── pkg/
│   └── llm/             — LLM API 调用封装
├── data/                 — CSV 数据目录
└── migrations/           — SQLite Schema
```

---

## 2. 核心数据流

### 2.1 风格分析流程（全 Go）

```
用户输入（历史交易记录 JSON）
    ↓
Gin POST /api/style/analyze
    ↓
StyleHandler.Analyze()
    ↓ Go HTTP Client → LLM API（OpenAI/Groq）
    ↓
解析 JSON 响应 → StyleProfile
    ↓
返回：风格画像
```

### 2.2 策略生成流程（全 Go）

```
用户输入（风格画像）
    ↓
Gin POST /api/strategy/generate
    ↓
StrategyHandler.Generate()
    ↓ Go HTTP Client → LLM API（策略生成 Prompt）
    ↓
解析 JSON 响应 → StrategyRules
    ↓
in-process call: BacktestEngine.Run(rules, data)
    ↓
返回：StrategyReport（策略 + 回测绩效）
```

### 2.3 回测流程（Go）

```
读取 CSV 数据（K线）
    ↓
StrategyExecutor.Init(rules) — 初始化指标
    ↓
逐 Bar 遍历
    ↓
信号生成（买入/卖出/持有）
    ↓
撮合（按收盘价成交）
    ↓
更新账户/持仓
    ↓
计算绩效指标
    ↓
返回 Report
```

---

## 3. API 接口定义（Gin HTTP）

所有接口均为 Go HTTP Handler，无外部调用开销。

| 方法 | 路径 | 描述 | 请求体 | 响应 |
|------|------|------|--------|------|
| POST | `/api/style/analyze` | 分析用户风格 | `{ records: [...] }` | `StyleProfile` |
| POST | `/api/strategy/generate` | 生成策略 | `{ styleProfile: {...} }` | `StrategyReport` |
| GET | `/api/strategies` | 策略列表 | - | `Strategy[]` |
| GET | `/api/strategies/:id` | 策略详情 | - | `Strategy` |
| POST | `/api/backtest` | 执行回测 | `{ strategyId, symbol, days }` | `BacktestReport` |
| GET | `/api/backtest/:jobId` | 回测结果 | - | `BacktestReport` |
| DELETE | `/api/strategies/:id` | 删除策略 | - | - |

**注意**：所有 Handler 均为 in-process Go 调用，无网络开销。LLM 调用仅在 AI Engine 模块中通过 HTTP Client 访问外部 API。

---

## 4. 数据存储

### 4.1 SQLite（Go 侧）

```sql
-- 策略表
CREATE TABLE strategies (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  style TEXT NOT NULL,
  rules JSON NOT NULL,
  version INTEGER DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 回测历史表
CREATE TABLE backtest_results (
  id TEXT PRIMARY KEY,
  strategy_id TEXT REFERENCES strategies(id),
  symbol TEXT NOT NULL,
  days INTEGER,
  metrics JSON NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 4.2 CSV 数据格式

```csv
date,open,high,low,close,volume
2025-01-02,10.50,10.80,10.40,10.75,1000000
```

---

## 5. AI Prompt 设计

### 5.1 风格分析 Prompt

```
你是一个量化投资风格分析师。根据用户的历史交易记录，
分析出以下风格指标：
- risk_score: 0-100 的风险评分
- style: 保守/稳健/平衡/积极/激进
- trade_frequency: 高频/中频/低频
- avg_hold_days: 平均持仓天数
- max_drawdown_tolerance: 可接受最大回撤百分比

输出 JSON 格式。
```

### 5.2 策略生成 Prompt

```
你是一个量化策略专家。根据用户的风格画像，
生成一个量化交易策略规则。

风格画像：{style_profile}

策略规则必须包含：
1. indicators: 使用的指标组合（MA/RSI/MACD/布林带）
2. params: 每个指标的具体参数
3. entry: 入场条件
4. exit: 出场条件（止盈/止损）
5. position: 仓位管理规则

输出严格 JSON 格式。
```

---

## 6. 项目计划

| 里程碑 | 内容 | 产出 | 预计 |
|--------|------|------|------|
| M1 | 架构搭建 + 风格分析 + 策略生成 + 回测引擎 | 可演示的 MVP | 1周 |
| M2 | 策略库 CRUD + 回测历史 + 对比界面 | 完整前端 | 1周 |
| M3 | 参数调优 + 多市场数据 | 功能完善 | 1周 |
| M4 | 部署上线 | 可访问 URL | 1周 |
| M5 | 验收 | 交付报告 | — |

### M1 详细任务

| Task | 描述 |
|------|------|
| T1 | 项目初始化（单一 Go 模块）|
| T2 | 数据模型定义（Go struct）|
| T3 | CSV 数据加载器 |
| T4 | 指标计算（MA/RSI/MACD/布林带）|
| T5 | 回测引擎核心 |
| T6 | 绩效指标计算 |
| T7 | Gin HTTP 接口 |
| T8 | AI Client（Go HTTP → LLM API）|
| T9 | 风格分析 Handler |
| T10 | 策略生成 Handler |
| T11 | SQLite 存储层 |
| T12 | 基础前端界面 |
| T13 | TDD 测试 + GAN 评审 |

---

## 7. 技术决策记录（ADR）

| ID | 决策 | 理由 |
|----|------|------|
| ADR-001 | **全 Go** | 单一二进制、部署简单、无 Python 运行时 |
| ADR-002 | **Gin HTTP** | 轻量、高性能、成熟 |
| ADR-003 | **Go HTTP Client 调用 LLM** | 直接调用 OpenAI/Groq API，无中间层 |
| ADR-004 | **SQLite 存储** | 简单、无外部依赖、内嵌 |
| ADR-005 | **React 前端** | 交互友好 |
