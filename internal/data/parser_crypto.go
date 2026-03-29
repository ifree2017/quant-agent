package data

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"
)

// loadCrypto 解析数字货币CSV
// 格式：date,open,high,low,close,volume,quote_volume
// 注意：币圈通常用quote_volume（成交额）代替成交量
func (l *LoaderV2) loadCrypto(filename string, days int) ([]Bar, MarketType, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, "", err
	}

	bars := make([]Bar, 0, len(records)-1)
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 5 {
			continue
		}
		date, _ := time.Parse("2006-01-02", strings.TrimSpace(record[0]))
		open, _ := strconv.ParseFloat(strings.TrimSpace(record[1]), 64)
		high, _ := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
		low, _ := strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
		close, _ := strconv.ParseFloat(strings.TrimSpace(record[4]), 64)

		// volume字段可能是第6列(quote_volume)或第5列(volume)
		var volume int64
		if len(record) >= 6 {
			volume, _ = strconv.ParseInt(strings.TrimSpace(record[5]), 10, 64)
		} else {
			volume, _ = strconv.ParseInt(strings.TrimSpace(record[4]), 10, 64)
		}

		bars = append(bars, Bar{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	if len(bars) > days && days > 0 {
		bars = bars[len(bars)-days:]
	}

	return bars, MarketCrypto, nil
}
