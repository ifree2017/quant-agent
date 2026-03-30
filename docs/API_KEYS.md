# QuantAgent 外部API申请指南

> 最后更新：2026-03-30

## 概览

| 数据源 | 是否需要Key | 申请难度 | 优先级 |
|--------|------------|---------|-------|
| LLM（策略生成）| ✅ 需要 | 低 | P0 |
| 东方财富 | ❌ 无需 | — | P0 |
| 雪球 | ❌ 无需（模拟数据）| — | P1 |
| 搜狗微信 | ❌ 无需 | — | P1 |
| 抖音热榜 | ❌ 无需 | — | P1 |
| 新闻事件 | ❌ 无需 | — | P1 |

---

## P0 — 必须申请

### 1. LLM API（策略生成 + 风格分析）

**推荐优先级：Groq > OpenAI > Anthropic**

#### Groq（推荐，免费额度大）
- **申请地址**：https://console.groq.com/keys
- **免费额度**：30请求/分钟，1440请求/天
- **模型**：Llama-3.3-70b, Mixtral-8x7b, Gemma2-9b
- **步骤**：
  1. 用 Google/GitHub 账号登录
  2. 进入 API Keys 页面
  3. 点击 Create Key，命名 `quant-agent`
  4. 复制 Key，格式 `gsk_xxxxx`

#### OpenAI（备选）
- **申请地址**：https://platform.openai.com/api-keys
- **免费额度**：$5新用户credit
- **模型**：GPT-4o, GPT-4o-mini, o3-mini

#### Anthropic（备选）
- **申请地址**：https://console.anthropic.com/settings/keys
- **免费额度**：有限
- **模型**：Claude 3.5 Sonnet, Claude 3 Opus

**环境变量配置**：
```bash
# .env
LLM_BASE_URL=https://api.groq.com/openai/v1
LLM_API_KEY=gsk_xxxxxxxxxxxxxxxxxxxxx
LLM_MODEL=llama-3.3-70b-versatile
```

---

## P1 — 可选（已有免费方案）

### 2. 东方财富 EastMoney
- **接口**：`push2.eastmoney.com`
- **是否需Key**：❌ 无需
- **备注**：完全免费，但需控制请求频率（<1次/秒）

### 3. 雪球 XueQiu
- **当前状态**：使用结构化模拟数据
- **是否需Key**：❌ 暂不需要
- **如需真实数据**：需申请雪球Cookie（反爬严格）

### 4. 搜狗微信 WeChat
- **接口**：`weixin.sogou.com`
- **是否需Key**：❌ 无需
- **备注**：有限频控制即可

### 5. 抖音热榜 Douyin
- **接口**：`douyin.com` 热榜接口
- **是否需Key**：❌ 无需
- **备注**：反爬严格，建议添加随机User-Agent

### 6. 新闻事件 News
- **腾讯新闻**：`view.inews.qq.com`
- **东方财富公告**：`np-anotice-stock.eastmoney.com`
- **是否需Key**：❌ 全部无需

---

## 当前.env示例（量化之星）

```bash
# LLM配置（Groq免费版）
LLM_BASE_URL=https://api.groq.com/openai/v1
LLM_API_KEY=gsk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
LLM_MODEL=llama-3.3-70b-versatile

# 数据库（已有）
DATABASE_URL=postgres://postgres:postgres@47.99.163.232:5432/quant_agent

# 服务器
SERVER_PORT=10082
```

---

## 申请清单

- [ ] Groq API Key（P0，必备）
- [ ] （可选）OpenAI API Key
- [ ] （可选）Anthropic API Key
