package data

import (
	"os"
	"path/filepath"
	"testing"
)

func testDataDirV2(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	dataDir := filepath.Join(projectRoot, "data")
	return dataDir
}

func TestLoaderV2_DetectMarket_Default(t *testing.T) {
	l := NewLoaderV2("./data")
	// 未知类型默认返回A股
	got := l.detectMarket("unknown_symbol")
	if got != MarketAShare {
		t.Errorf("detectMarket(unknown): got %s, want a_share", got)
	}
}

func TestLoaderV2_FindFile(t *testing.T) {
	dataDir := testDataDirV2(t)
	l := NewLoaderV2(dataDir)

	// 000001.csv 应该存在于 data/
	path := l.findFile("000001", MarketAShare)
	if path == "" {
		t.Error("findFile should find 000001.csv")
	}
}

func TestLoaderV2_FindFile_Crypto(t *testing.T) {
	dataDir := testDataDirV2(t)
	l := NewLoaderV2(dataDir)

	// BTC 应该存在于 data/crypto/BTC.csv
	path := l.findFile("BTC", MarketCrypto)
	if path == "" {
		t.Error("findFile should find BTC.csv in crypto dir")
	}
}

func TestLoaderV2_FindFile_NotFound(t *testing.T) {
	l := NewLoaderV2("./nonexistent")
	path := l.findFile("NONEXISTENT", MarketAShare)
	if path != "" {
		t.Errorf("findFile should return empty for nonexistent file, got %s", path)
	}
}

func TestLoaderV2_LoadBarsAdvanced(t *testing.T) {
	dataDir := testDataDirV2(t)
	l := NewLoaderV2(dataDir)

	bars, market, err := l.LoadBarsAdvanced("000001", 10)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced: %v", err)
	}
	if market != MarketAShare {
		t.Errorf("market: got %s, want a_share", market)
	}
	if len(bars) == 0 {
		t.Error("bars should not be empty")
	}
	if len(bars) > 10 {
		t.Errorf("bars length: got %d, want <= 10", len(bars))
	}
}

func TestLoaderV2_LoadBarsAdvanced_Crypto(t *testing.T) {
	dataDir := testDataDirV2(t)
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
	// BTC价格应该在合理范围 (数据中是97000左右)
	if bars[0].Close < 1000 {
		t.Errorf("BTC close price seems wrong: %.2f", bars[0].Close)
	}
}

func TestLoaderV2_LoadBarsAdvanced_FileNotFound(t *testing.T) {
	l := NewLoaderV2("./data")
	_, _, err := l.LoadBarsAdvanced("NONEXISTENT", 10)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoaderV2_LoadBarsAdvanced_DaysLimit(t *testing.T) {
	dataDir := testDataDirV2(t)
	l := NewLoaderV2(dataDir)

	// 请求5天，但数据可能有更多
	bars, _, err := l.LoadBarsAdvanced("000001", 5)
	if err != nil {
		t.Fatalf("LoadBarsAdvanced: %v", err)
	}
	if len(bars) > 5 {
		t.Errorf("bars should be limited to 5, got %d", len(bars))
	}
}
