package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // 分あたりのリクエスト数
	window   time.Duration // 時間窓
}

type Visitor struct {
	requests []time.Time
	mu       sync.Mutex
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     requestsPerMinute,
		window:   time.Minute,
	}

	// 古いエントリを定期的にクリーンアップ
	go rl.cleanupVisitors()

	return rl
}

// 指定されたIPアドレスのリクエストを許可するかチェック
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.RLock()
	visitor, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		visitor = &Visitor{
			requests: make([]time.Time, 0),
		}
		rl.visitors[ip] = visitor
		rl.mu.Unlock()
	}

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// 古いリクエストを削除
	validRequests := make([]time.Time, 0)
	for _, requestTime := range visitor.requests {
		if requestTime.After(cutoff) {
			validRequests = append(validRequests, requestTime)
		}
	}
	visitor.requests = validRequests

	// レート制限チェック
	if len(visitor.requests) >= rl.rate {
		return false
	}

	// リクエストを記録
	visitor.requests = append(visitor.requests, now)
	return true
}

// cleanupVisitors 古い訪問者データを定期的にクリーンアップ
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-rl.window * 2) // 余裕を持って削除

			for ip, visitor := range rl.visitors {
				visitor.mu.Lock()
				hasRecentRequests := false
				for _, requestTime := range visitor.requests {
					if requestTime.After(cutoff) {
						hasRecentRequests = true
						break
					}
				}
				visitor.mu.Unlock()

				if !hasRecentRequests {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// RateLimitConfig レート制限の設定
type RateLimitConfig struct {
	General int // 一般的なエンドポイントのレート制限（分間）
	Auth    int // 認証関連エンドポイントのレート制限（分間）
	Contact int // お問合せエンドポイントのレート制限（分間）
}

// DefaultRateLimitConfig 環境変数からレート制限設定を取得
func DefaultRateLimitConfig() *RateLimitConfig {
	config := &RateLimitConfig{
		General: 120, // デフォルト: 120req/min
		Auth:    20,  // デフォルト: 20req/min
		Contact: 3,   // デフォルト: 3req/min（お問合せは厳しめ）
	}

	// 環境変数から設定を取得
	if generalStr := os.Getenv("RATE_LIMIT_GENERAL"); generalStr != "" {
		if general, err := strconv.Atoi(generalStr); err == nil && general > 0 {
			config.General = general
		}
	}

	if authStr := os.Getenv("RATE_LIMIT_AUTH"); authStr != "" {
		if auth, err := strconv.Atoi(authStr); err == nil && auth > 0 {
			config.Auth = auth
		}
	}

	if contactStr := os.Getenv("RATE_LIMIT_CONTACT"); contactStr != "" {
		if contact, err := strconv.Atoi(contactStr); err == nil && contact > 0 {
			config.Contact = contact
		}
	}

	return config
}

// RateLimitMiddleware レート制限ミドルウェア
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !limiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "300")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"リクエストが多すぎます。しばらく待ってから再試行してください。"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthRateLimitMiddleware 認証エンドポイント用のレート制限ミドルウェア
func AuthRateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !limiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "300")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"認証試行回数が上限に達しました。しばらく待ってから再試行してください。"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ContactRateLimitMiddleware お問合せエンドポイント用のレート制限ミドルウェア
func ContactRateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !limiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "300")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"お問い合わせの送信回数が上限に達しました。しばらく待ってから再試行してください。"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP クライアントのIPアドレスを取得
func getClientIP(r *http.Request) string {
	// プロキシヘッダーをチェック
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// 最初のIPアドレスを取得（クライアントIP）
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		if net.ParseIP(realIP) != nil {
			return realIP
		}
	}

	// リモートアドレスから取得
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// RateLimitResponse レート制限情報をレスポンスヘッダーに追加
func (rl *RateLimiter) AddHeaders(w http.ResponseWriter, ip string) {
	rl.mu.RLock()
	visitor, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if !exists {
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.rate))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rl.rate))
		return
	}

	visitor.mu.Lock()
	currentRequests := len(visitor.requests)
	visitor.mu.Unlock()

	remaining := rl.rate - currentRequests
	if remaining < 0 {
		remaining = 0
	}

	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.rate))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rl.window).Unix()))
}
