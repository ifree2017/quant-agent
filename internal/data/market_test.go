package data

import "testing"

func TestDetectMarket(t *testing.T) {
	loader := NewLoaderV2("./data")

	tests := []struct {
		symbol string
		want   MarketType
	}{
		{"000001", MarketAShare},
		{"600519", MarketAShare},
		{"rb2101", MarketFuture},
		{"BTC", MarketCrypto},
		{"ETH", MarketCrypto},
	}

	for _, tt := range tests {
		got := loader.detectMarket(tt.symbol)
		if got != tt.want {
			t.Errorf("detectMarket(%s): got %s, want %s", tt.symbol, got, tt.want)
		}
	}
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"000001", true},
		{"600519", true},
		{"123456", true},
		{"rb2101", false},
		{"BTC", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isDigit(tt.s)
		if got != tt.want {
			t.Errorf("isDigit(%s): got %v, want %v", tt.s, got, tt.want)
		}
	}
}
