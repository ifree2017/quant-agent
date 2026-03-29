# quant-agent — STATE.md

## 项目信息
- **项目名**：QuantAgent 智能量化策略系统
- **目录**：`projects/quant-agent`
- **状态**：`SDD_IN_PROGRESS`
- **创建时间**：2026-03-29
- **最后更新**：2026-03-29

---

## SDD 文档

| 文档 | 路径 | 状态 |
|------|------|------|
| 需求规格说明书 | `docs/requirements/QuantAgent_需求规格说明书.md` | ✅ 已创建 |
| 技术方案 | `docs/design/QuantAgent_技术方案.md` | ✅ 已创建 |
| 接口评审记录 | `docs/review/QuantAgent_接口评审记录.md` | ✅ 已创建，待三方签字 |

---

## 里程碑（M1-M5）

| 里程碑 | 状态 | 完成日期 |
|--------|------|---------|
| SDD 评审 | ✅ 通过（v1.2） | 2026-03-29 |
| M1 | 🔄 进行中 | — |
| M2 | ⏳ 待开始 | — |
| M3 | ⏳ 待开始 | — |
| M4 | ⏳ 待开始 | — |
| M5 | ⏳ 待开始 | — |

---

## 技术栈

| 模块 | 技术 |
|------|------|
| API 层 | Python / FastAPI |
| AI 引擎 | OpenAI / Groq API |
| 回测引擎 | Go |
| 数据存储 | SQLite |
| 前端 | React / TypeScript |
| 通信 | gRPC |

---

## 核心功能

1. 用户操作风格分析（AI）
2. AI 策略生成
3. Go 回测引擎
4. 绩效指标报告
5. 策略库管理（M2）
6. 策略对比（M2）

---

## 项目结构

```
quant-agent/
├── docs/
│   ├── requirements/    # 需求规格说明书
│   ├── design/         # 技术方案
│   └── review/         # 接口评审记录
├── python/             # Python 层（API + AI）
│   ├── api/
│   ├── ai/
│   └── main.py
├── go/                 # Go 层（回测核心）
│   ├── cmd/
│   └── internal/
├── frontend/           # React 前端
└── STATE.md
```

---

## 下一步

1. SDD 三方签字评审
2. 评审通过后 → M1 启动
