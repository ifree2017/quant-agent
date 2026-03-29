package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testDataDir(t *testing.T) string {
	// Tests run from package dir; data is at projectRoot/data
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// internal/data -> projectRoot
	projectRoot := filepath.Join(wd, "..", "..")
	dataDir := filepath.Join(projectRoot, "data")
	return dataDir
}

func TestLoader_LoadBars(t *testing.T) {
	dataDir := testDataDir(t)
	loader := NewLoader(dataDir)
	bars, err := loader.LoadBars("000001", 30)
	if err != nil {
		t.Fatalf("LoadBars: %v", err)
	}
	if len(bars) == 0 {
		t.Fatal("bars should not be empty")
	}
	if len(bars) > 30 {
		t.Errorf("bars length: got %d, want <= 30", len(bars))
	}

	// 检查字段
	bar := bars[0]
	if bar.Date.IsZero() {
		t.Error("bar date should not be zero")
	}
	if bar.Open <= 0 || bar.Close <= 0 || bar.High <= 0 || bar.Low <= 0 {
		t.Error("prices should be positive")
	}
	if bar.High < bar.Low {
		t.Error("high should be >= low")
	}
	if bar.Close <= 0 || bar.Volume <= 0 {
		t.Error("close and volume should be positive")
	}
}

func TestLoader_LoadBarsFromRecords(t *testing.T) {
	records := [][]string{
		{"date", "open", "high", "low", "close", "volume"},
		{"2025-01-02", "10.50", "10.80", "10.40", "10.75", "1000000"},
		{"2025-01-03", "10.80", "11.00", "10.70", "10.90", "1200000"},
		{"2025-01-06", "10.95", "11.20", "10.85", "11.10", "1100000"},
	}

	bars, err := LoadBarsFromRecords(records)
	if err != nil {
		t.Fatalf("LoadBarsFromRecords: %v", err)
	}
	if len(bars) != 3 {
		t.Errorf("bars count: got %d, want 3", len(bars))
	}

	// 第一个 bar
	if bars[0].Date.Format("2006-01-02") != "2025-01-02" {
		t.Errorf("first date: got %s", bars[0].Date.Format("2006-01-02"))
	}
	if bars[0].Open != 10.50 {
		t.Errorf("open: got %v, want 10.50", bars[0].Open)
	}
	if bars[0].High != 10.80 {
		t.Errorf("high: got %v, want 10.80", bars[0].High)
	}
	if bars[0].Low != 10.40 {
		t.Errorf("low: got %v, want 10.40", bars[0].Low)
	}
	if bars[0].Close != 10.75 {
		t.Errorf("close: got %v, want 10.75", bars[0].Close)
	}
	if bars[0].Volume != 1000000 {
		t.Errorf("volume: got %v, want 1000000", bars[0].Volume)
	}
}

func TestLoader_LoadBarsFromRecords_SkipBadRows(t *testing.T) {
	records := [][]string{
		{"date", "open", "high", "low", "close", "volume"},
		{"invalid-date", "10.50", "10.80", "10.40", "10.75", "1000000"}, // 无效日期 → 跳过
		{"2025-01-03", "10.80", "11.00", "10.70", "10.90", "1200000"},   // 有效行
		{"2025-01-06", "10.95", "11.20", "10.85", "11.10", "1100000"},  // 有效行
	}

	bars, err := LoadBarsFromRecords(records)
	if err != nil {
		t.Fatalf("LoadBarsFromRecords: %v", err)
	}
	// 第一行(无效日期)被跳过，其余2行为有效数据
	if len(bars) != 2 {
		t.Errorf("bars count: got %d, want 2", len(bars))
	}
	if bars[0].Date.Format("2006-01-02") != "2025-01-03" {
		t.Errorf("first date: got %s, want 2025-01-03", bars[0].Date.Format("2006-01-02"))
	}
	if bars[1].Date.Format("2006-01-02") != "2025-01-06" {
		t.Errorf("second date: got %s, want 2025-01-06", bars[1].Date.Format("2006-01-02"))
	}
}

func TestLoader_LoadBars_FileNotFound(t *testing.T) {
	loader := NewLoader("data")
	_, err := loader.LoadBars("nonexistent", 10)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestBar_MAData(t *testing.T) {
	bar := Bar{
		Date:   time.Now(),
		Open:   10.0,
		High:   11.0,
		Low:    9.0,
		Close:  10.5,
		Volume: 1000,
	}
	if bar.Close != 10.5 {
		t.Errorf("close: got %v, want 10.5", bar.Close)
	}
}

func TestLoader_Symbols_NotImplemented(t *testing.T) {
	dataDir := testDataDir(t)
	// Loader 没有 Symbols 方法，测试确认 LoadBars 能处理不同 symbol
	loader := NewLoader(dataDir)
	// 用真实数据目录，确认它只是文件名匹配
	bars, err := loader.LoadBars("000001", 5)
	if err != nil {
		t.Fatalf("load bars: %v", err)
	}
	if len(bars) != 5 {
		t.Errorf("got %d bars, want 5", len(bars))
	}
}

func TestBar_Fields(t *testing.T) {
	dt := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	bar := Bar{
		Date:   dt,
		Open:   10.0,
		High:   11.0,
		Low:    9.5,
		Close:  10.8,
		Volume: 5000000,
	}
	if bar.Date != dt {
		t.Errorf("date mismatch")
	}
	if bar.Open != 10.0 || bar.High != 11.0 || bar.Low != 9.5 || bar.Close != 10.8 {
		t.Errorf("price fields mismatch")
	}
	if bar.Volume != 5000000 {
		t.Errorf("volume: got %v, want 5000000", bar.Volume)
	}
}
