# QuantAgent — 智能量化策略系统

> AI 驱动量化投研平台：分析交易风格 → 生成策略 → 回测验证 → 绩效评估

## 功能

- **🧠 风格分析**：基于历史交易记录，AI 分析用户投资风格（保守/稳健/平衡/积极/激进）
- **📊 策略生成**：根据风格画像，AI 自动生成量化交易策略规则
- **⚡ 回测引擎**：Go 高性能回测引擎，支持 MA/RSI/MACD/Bollinger 指标
- **📈 绩效评估**：夏普比率、最大回撤、胜率、盈亏比等核心指标
- **💾 策略管理**：PostgreSQL 存储策略和回测历史

## 技术栈

| 模块 | 技术 |
|------|------|
| 语言 | Go 1.24 |
| HTTP API | Gin |
| AI 引擎 | OpenAI / Groq API（Go HTTP Client）|
| 数据存储 | PostgreSQL |
| 前端 | React + TypeScript + Vite |
| 单元测试 | Go testing |

## 快速启动

### 前置条件

- Go 1.24+
- Node.js 18+
- PostgreSQL 14+

### 1. 克隆

```bash
git clone https://github.com/ifree2017/quant-agent.git
cd quant-agent
```

### 2. 数据库

```bash
# 创建数据库（PostgreSQL）
psql -h 47.99.163.232 -U postgres -c "CREATE DATABASE quant_agent;"

# 运行 Migration
go run cmd/migrate/main.go
```

### 3. 配置

```bash
cp .env.example .env
# 编辑 .env，填入 LLM API Key
```

`.env` 示例：

```env
LLM_URL=https://api.openai.com/v1
LLM_TOKEN=your_api_key_here
SERVER_ADDR=:8080
DATA_DIR=./data
```

### 4. 启动后端

```bash
go run cmd/server/main.go
# 监听 :8080
```

### 5. 启动前端

```bash
cd frontend
npm install
npm run dev
# 访问 http://localhost:5173
```

## API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |
| POST | `/api/style/analyze` | 分析交易风格 |
| POST | `/api/strategy/generate` | 生成策略 |
| GET | `/api/strategies` | 策略列表 |
| GET | `/api/strategies/:id` | 策略详情 |
| DELETE | `/api/strategies/:id` | 删除策略 |
| POST | `/api/backtest` | 执行回测 |
| GET | `/api/backtest/:id` | 回测结果 |

### 风格分析

```bash
curl -X POST http://localhost:8080/api/style/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "userID": "user1",
    "records": [
      {"date": "2025-01-01", "symbol": "000001", "action": "buy", "price": 10.0, "quantity": 100},
      {"date": "2025-01-15", "symbol": "000001", "action": "sell", "price": 11.0, "quantity": 100}
    ]
  }'
```

### 策略生成 + 回测

```bash
curl -X POST http://localhost:8080/api/strategy/generate \
  -H "Content-Type: application/json" \
  -d '{
    "userID": "user1",
    "styleProfile": {"style": "稳健", "riskScore": 45},
    "symbol": "000001"
  }'
```

## 项目结构

```
quant-agent/
├── cmd/
│   ├── server/          # HTTP 服务器入口
│   └── migrate/         # 数据库 Migration 工具
├── internal/
│   ├── ai/              # LLM API Client
│   ├── api/             # Gin HTTP Handler
│   ├── backtest/        # 回测引擎
│   ├── data/            # CSV 数据加载
│   ├── model/           # 数据模型
│   ├── report/          # 绩效指标计算
│   ├── strategy/        # 策略执行器（MA/RSI/MACD/Bollinger）
│   └── store/           # PostgreSQL 存储层
├── migrations/           # SQL Migration
├── frontend/             # React 前端
├── data/                 # CSV 历史数据
└── docs/                 # SDD 文档
```

## 回测指标

| 指标 | 说明 |
|------|------|
| Total Return | 总收益率 |
| Sharpe Ratio | 年化夏普比率 |
| Max Drawdown | 最大回撤 |
| Win Rate | 胜率 |
| Profit Loss Ratio | 盈亏比 |
| Calmar Ratio | 卡玛比率 |
| Sortino Ratio | 索提诺比率 |
| Annual Return | 年化收益率 |

## 开发

```bash
# 运行测试
go test ./...

# 构建
go build ./...

# 前端构建
cd frontend && npm run build
```

## License

MIT
