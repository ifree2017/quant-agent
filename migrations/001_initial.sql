-- 用户风格表
CREATE TABLE IF NOT EXISTS style_profiles (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  style TEXT NOT NULL,
  risk_score INTEGER,
  trade_frequency TEXT,
  avg_hold_days INTEGER,
  max_drawdown_tolerance REAL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 策略表
CREATE TABLE IF NOT EXISTS strategies (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  style TEXT NOT NULL,
  rules JSONB NOT NULL,
  version INTEGER DEFAULT 1,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 回测结果表
CREATE TABLE IF NOT EXISTS backtest_results (
  id TEXT PRIMARY KEY,
  strategy_id TEXT REFERENCES strategies(id),
  symbol TEXT NOT NULL,
  days INTEGER,
  initial_cash REAL,
  final_cash REAL,
  metrics JSONB NOT NULL,
  trades JSONB,
  equity_curve JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_strategies_user ON strategies(user_id);
CREATE INDEX IF NOT EXISTS idx_backtest_strategy ON backtest_results(strategy_id);
CREATE INDEX IF NOT EXISTS idx_backtest_created ON backtest_results(created_at);
