package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Loader CSV数据加载器
type Loader struct {
	dataDir string
}

// NewLoader 创建加载器
func NewLoader(dataDir string) *Loader {
	return &Loader{dataDir: dataDir}
}

// LoadBars 加载K线数据
func (l *Loader) LoadBars(symbol string, days int) ([]Bar, error) {
	filename := filepath.Join(l.dataDir, symbol+".csv")
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open data file %s: %w", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	bars := make([]Bar, 0, len(records)-1)
	for i, row := range records {
		if i == 0 { // skip header
			continue
		}
		if len(row) < 6 {
			continue
		}
		date, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			continue
		}
		open, _ := strconv.ParseFloat(row[1], 64)
		high, _ := strconv.ParseFloat(row[2], 64)
		low, _ := strconv.ParseFloat(row[3], 64)
		close, _ := strconv.ParseFloat(row[4], 64)
		volume, _ := strconv.ParseInt(row[5], 10, 64)

		bars = append(bars, Bar{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	// 取最近days条
	if days > 0 && len(bars) > days {
		bars = bars[len(bars)-days:]
	}

	return bars, nil
}

// LoadBarsFromRecords 从内存数据加载（用于测试）
func LoadBarsFromRecords(records [][]string) ([]Bar, error) {
	bars := make([]Bar, 0, len(records))
	for i, row := range records {
		if i == 0 && row[0] == "date" { // skip header
			continue
		}
		if len(row) < 6 {
			continue
		}
		date, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			continue
		}
		open, _ := strconv.ParseFloat(row[1], 64)
		high, _ := strconv.ParseFloat(row[2], 64)
		low, _ := strconv.ParseFloat(row[3], 64)
		close, _ := strconv.ParseFloat(row[4], 64)
		volume, _ := strconv.ParseInt(row[5], 10, 64)

		bars = append(bars, Bar{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}
	return bars, nil
}
