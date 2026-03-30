package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"quant-agent/internal/model"
)

type Store struct {
	db *pgxpool.Pool
}

// NewStore 连接到 PostgreSQL，返回 Store 实例
func NewStore(ctx context.Context, connString string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conn string: %w", err)
	}
	cfg.MaxConns = 20

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &Store{db: pool}, nil
}

// Close closes the underlying connection pool.
func (s *Store) Close() {
	s.db.Close()
}

// DB returns the underlying connection pool for advanced operations.
func (s *Store) DB() *pgxpool.Pool {
	return s.db
}

// ---------------------------------------------------------------------------
// StyleProfile
// ---------------------------------------------------------------------------

// SaveStyleProfile inserts or updates a style profile.
func (s *Store) SaveStyleProfile(p model.StyleProfile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO style_profiles (id, user_id, style, risk_score, trade_frequency, avg_hold_days, max_drawdown_tolerance, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			style = EXCLUDED.style,
			risk_score = EXCLUDED.risk_score,
			trade_frequency = EXCLUDED.trade_frequency,
			avg_hold_days = EXCLUDED.avg_hold_days,
			max_drawdown_tolerance = EXCLUDED.max_drawdown_tolerance
	`, p.ID, p.UserID, p.Style, int(p.RiskScore), p.TradeFrequency, int(p.AvgHoldDays), p.MaxDrawdownTolerance, p.CreatedAt)

	return err
}

// GetStyleProfile retrieves a style profile by user ID.
func (s *Store) GetStyleProfile(userID string) (*model.StyleProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p model.StyleProfile
	var riskScore, avgHoldDays int
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, style, risk_score, trade_frequency, avg_hold_days, max_drawdown_tolerance, created_at
		FROM style_profiles WHERE user_id = $1
	`, userID).Scan(&p.ID, &p.UserID, &p.Style, &riskScore, &p.TradeFrequency, &avgHoldDays, &p.MaxDrawdownTolerance, &p.CreatedAt)

	if err != nil {
		return nil, err
	}
	p.RiskScore = float64(riskScore)
	p.AvgHoldDays = float64(avgHoldDays)
	return &p, nil
}

// ---------------------------------------------------------------------------
// Strategy
// ---------------------------------------------------------------------------

// SaveStrategy inserts or updates a strategy.
func (s *Store) SaveStrategy(strategy *model.Strategy) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	if strategy.CreatedAt.IsZero() {
		strategy.CreatedAt = now
	}
	if strategy.UpdatedAt.IsZero() {
		strategy.UpdatedAt = now
	}

	rulesJSON, err := json.Marshal(strategy.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO strategies (id, user_id, name, style, rules, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			style = EXCLUDED.style,
			rules = EXCLUDED.rules,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at
	`, strategy.ID, strategy.UserID, strategy.Name, strategy.Style, rulesJSON, strategy.Version, strategy.CreatedAt, strategy.UpdatedAt)

	return err
}

// GetStrategy retrieves a strategy by ID.
func (s *Store) GetStrategy(id string) (*model.Strategy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var strategy model.Strategy
	var rulesJSON []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, name, style, rules, version, created_at, updated_at
		FROM strategies WHERE id = $1
	`, id).Scan(&strategy.ID, &strategy.UserID, &strategy.Name, &strategy.Style, &rulesJSON, &strategy.Version, &strategy.CreatedAt, &strategy.UpdatedAt)

	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rulesJSON, &strategy.Rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}
	return &strategy, nil
}

// ListStrategies returns all strategies for a user.
func (s *Store) ListStrategies(userID string) ([]model.Strategy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, name, style, rules, version, created_at, updated_at
		FROM strategies WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strategies []model.Strategy
	for rows.Next() {
		var strategy model.Strategy
		var rulesJSON []byte
		if err := rows.Scan(&strategy.ID, &strategy.UserID, &strategy.Name, &strategy.Style, &rulesJSON, &strategy.Version, &strategy.CreatedAt, &strategy.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rulesJSON, &strategy.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}
		strategies = append(strategies, strategy)
	}
	return strategies, rows.Err()
}

// DeleteStrategy deletes a strategy by ID.
func (s *Store) DeleteStrategy(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `DELETE FROM strategies WHERE id = $1`, id)
	return err
}

// ---------------------------------------------------------------------------
// Backtest
// ---------------------------------------------------------------------------

// SaveBacktest inserts or replaces a backtest result.
func (s *Store) SaveBacktest(result *model.BacktestResult) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}

	metricsJSON, err := json.Marshal(result.Metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	tradesJSON, err := json.Marshal(result.Trades)
	if err != nil {
		return fmt.Errorf("failed to marshal trades: %w", err)
	}
	equityCurveJSON, err := json.Marshal(result.EquityCurve)
	if err != nil {
		return fmt.Errorf("failed to marshal equity curve: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO backtest_results (id, strategy_id, symbol, days, initial_cash, final_cash, metrics, trades, equity_curve, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			strategy_id = EXCLUDED.strategy_id,
			symbol = EXCLUDED.symbol,
			days = EXCLUDED.days,
			initial_cash = EXCLUDED.initial_cash,
			final_cash = EXCLUDED.final_cash,
			metrics = EXCLUDED.metrics,
			trades = EXCLUDED.trades,
			equity_curve = EXCLUDED.equity_curve
	`, result.ID, result.StrategyID, result.Symbol, result.Days, result.InitialCash, result.FinalCash, metricsJSON, tradesJSON, equityCurveJSON, result.CreatedAt)

	return err
}

// GetBacktest retrieves a backtest result by ID.
func (s *Store) GetBacktest(id string) (*model.BacktestResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result model.BacktestResult
	var metricsJSON, tradesJSON, equityCurveJSON []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, strategy_id, symbol, days, initial_cash, final_cash, metrics, trades, equity_curve, created_at
		FROM backtest_results WHERE id = $1
	`, id).Scan(&result.ID, &result.StrategyID, &result.Symbol, &result.Days, &result.InitialCash, &result.FinalCash, &metricsJSON, &tradesJSON, &equityCurveJSON, &result.CreatedAt)

	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(metricsJSON, &result.Metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}
	if err := json.Unmarshal(tradesJSON, &result.Trades); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trades: %w", err)
	}
	if err := json.Unmarshal(equityCurveJSON, &result.EquityCurve); err != nil {
		return nil, fmt.Errorf("failed to unmarshal equity curve: %w", err)
	}
	return &result, nil
}

// ListBacktests returns all backtest results for a strategy.
func (s *Store) ListBacktests(strategyID string) ([]model.BacktestResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `
		SELECT id, strategy_id, symbol, days, initial_cash, final_cash, metrics, trades, equity_curve, created_at
		FROM backtest_results WHERE strategy_id = $1 ORDER BY created_at DESC
	`, strategyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.BacktestResult
	for rows.Next() {
		var r model.BacktestResult
		var metricsJSON, tradesJSON, equityCurveJSON []byte
		if err := rows.Scan(&r.ID, &r.StrategyID, &r.Symbol, &r.Days, &r.InitialCash, &r.FinalCash, &metricsJSON, &tradesJSON, &equityCurveJSON, &r.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metricsJSON, &r.Metrics); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
		}
		if err := json.Unmarshal(tradesJSON, &r.Trades); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trades: %w", err)
		}
		if err := json.Unmarshal(equityCurveJSON, &r.EquityCurve); err != nil {
			return nil, fmt.Errorf("failed to unmarshal equity curve: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
