package report

import (
	"testing"
	"time"
	"quant-agent/internal/model"
	"quant-agent/internal/data"
)

func testDate(days int) time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, days)
}

func TestMetrics_CalcMetrics(t *testing.T) {
	trades := []model.Trade{
		{EntryDate: "2025-01-01", ExitDate: "2025-01-15", EntryPrice: 10.0, ExitPrice: 11.0, Quantity: 100, PnL: 100.0},
		{EntryDate: "2025-01-16", ExitDate: "2025-01-30", EntryPrice: 11.0, ExitPrice: 10.5, Quantity: 100, PnL: -50.0},
		{EntryDate: "2025-01-31", ExitDate: "2025-02-10", EntryPrice: 10.5, ExitPrice: 12.0, Quantity: 100, PnL: 150.0},
	}

	bars := []data.Bar{
		{Date: testDate(0), Close: 10.0},
		{Date: testDate(14), Close: 11.0},
	}

	metrics := CalcMetrics(trades, 10000.0, 10200.0, bars)

	if metrics.TotalTrades != 3 {
		t.Errorf("TotalTrades: got %d, want 3", metrics.TotalTrades)
	}
	if metrics.WinRate < 0.66 || metrics.WinRate > 0.68 {
		t.Errorf("WinRate: got %.2f, want ~0.67", metrics.WinRate)
	}
	if metrics.ProfitLossRatio < 2.4 || metrics.ProfitLossRatio > 2.6 {
		t.Errorf("ProfitLossRatio: got %.2f, want ~2.5", metrics.ProfitLossRatio)
	}
}
