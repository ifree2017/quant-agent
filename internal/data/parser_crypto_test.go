package data

import (
	"os"
	"path/filepath"
	"testing"
)

func testDataDirForCrypto(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	return filepath.Join(projectRoot, "data")
}

func TestLoadCrypto(t *testing.T) {
	dataDir := testDataDirForCrypto(t)
	l := NewLoaderV2(dataDir)
	bars, market, err := l.LoadBarsAdvanced("BTC", 5)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced(BTC): %v", err)
	}
	if market != MarketCrypto {
		t.Errorf("market: got %s, want crypto", market)
	}
	if len(bars) == 0 {
		t.Error("bars should not be empty for BTC")
	}
	// BTC价格应该在合理范围
	if bars[0].Close < 1000 {
		t.Errorf("BTC close price seems wrong: %.2f", bars[0].Close)
	}
}

func TestLoadCrypto_DaysLimit(t *testing.T) {
	dataDir := testDataDirForCrypto(t)
	l := NewLoaderV2(dataDir)
	bars, _, err := l.LoadBarsAdvanced("BTC", 2)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced(BTC): %v", err)
	}
	if len(bars) > 2 {
		t.Errorf("bars should be limited to 2, got %d", len(bars))
	}
}

func TestLoadCrypto_AllData(t *testing.T) {
	dataDir := testDataDirForCrypto(t)
	l := NewLoaderV2(dataDir)
	// 请求超过实际天数的days=0表示全部
	bars, _, err := l.LoadBarsAdvanced("BTC", 0)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced(BTC): %v", err)
	}
	if len(bars) < 5 {
		t.Errorf("should load all BTC bars, got %d", len(bars))
	}
}

func TestLoadCrypto_Fields(t *testing.T) {
	dataDir := testDataDirForCrypto(t)
	l := NewLoaderV2(dataDir)
	bars, _, err := l.LoadBarsAdvanced("BTC", 1)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced(BTC): %v", err)
	}
	if len(bars) == 0 {
		t.Fatal("bars should not be empty")
	}

	bar := bars[0]
	if bar.Open <= 0 {
		t.Errorf("open should be positive, got %.2f", bar.Open)
	}
	if bar.High <= 0 {
		t.Errorf("high should be positive, got %.2f", bar.High)
	}
	if bar.Low <= 0 {
		t.Errorf("low should be positive, got %.2f", bar.Low)
	}
	if bar.Close <= 0 {
		t.Errorf("close should be positive, got %.2f", bar.Close)
	}
	if bar.Volume <= 0 {
		t.Errorf("volume should be positive, got %d", bar.Volume)
	}
	if bar.High < bar.Low {
		t.Error("high should be >= low")
	}
}

func TestLoadFuture(t *testing.T) {
	dataDir := testDataDirForCrypto(t)
	l := NewLoaderV2(dataDir)
	bars, market, err := l.LoadBarsAdvanced("rb2101", 5)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced(rb2101): %v", err)
	}
	if market != MarketFuture {
		t.Errorf("market: got %s, want future", market)
	}
	if len(bars) == 0 {
		t.Error("bars should not be empty for rb2101")
	}
	// 期货价格应该在合理范围
	if bars[0].Close < 1000 {
		t.Errorf("rb2101 close price seems wrong: %.2f", bars[0].Close)
	}
}

func TestLoadFuture_DaysLimit(t *testing.T) {
	dataDir := testDataDirForCrypto(t)
	l := NewLoaderV2(dataDir)
	bars, _, err := l.LoadBarsAdvanced("rb2101", 2)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced(rb2101): %v", err)
	}
	if len(bars) > 2 {
		t.Errorf("bars should be limited to 2, got %d", len(bars))
	}
}

func TestLoadFuture_Fields(t *testing.T) {
	dataDir := testDataDirForCrypto(t)
	l := NewLoaderV2(dataDir)
	bars, _, err := l.LoadBarsAdvanced("rb2101", 1)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced(rb2101): %v", err)
	}
	if len(bars) == 0 {
		t.Fatal("bars should not be empty")
	}

	bar := bars[0]
	if bar.Open <= 0 {
		t.Errorf("open should be positive, got %.2f", bar.Open)
	}
	if bar.High < bar.Low {
		t.Error("high should be >= low")
	}
}
