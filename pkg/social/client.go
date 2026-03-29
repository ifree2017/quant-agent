package social

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter 限速器：每秒最多1个请求
type RateLimiter struct {
	mu       sync.Mutex
	lastTime time.Time
	interval time.Duration
}

// NewRateLimiter 创建限速器，interval 通常为 1 秒
func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		interval: interval,
	}
}

// Wait 等待直到可以发起请求
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(r.lastTime)
	if elapsed < r.interval {
		time.Sleep(r.interval - elapsed)
	}
	r.lastTime = time.Now()
}

// Client HTTP客户端（带限速）
type Client struct {
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// NewClient 创建 social API 客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		rateLimiter: NewRateLimiter(time.Second), // 每秒最多1个请求
	}
}

// DoRequest 执行HTTP请求（带限速）
func (c *Client) DoRequest(req *http.Request) (*http.Response, error) {
	c.rateLimiter.Wait()
	return c.httpClient.Do(req)
}

// Get 发起GET请求
func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.eastmoney.com")
	return c.DoRequest(req)
}
