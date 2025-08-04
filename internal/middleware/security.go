package middleware

import (
	"net/http"
	"os"
	"strings"
)

// SecurityConfig セキュリティミドルウェアの設定
type SecurityConfig struct {
	SupabaseURL    string
	Environment    string
	AllowedDomains []string
	EnableHSTS     bool
	EnableCSP      bool
}

// NewSecurityConfig 環境変数からセキュリティ設定を作成
func NewSecurityConfig() *SecurityConfig {
	config := &SecurityConfig{
		SupabaseURL: os.Getenv("SUPABASE_URL"),
		Environment: getEnv("ENV", "development"),
		EnableHSTS:  getBoolEnv("SECURITY_ENABLE_HSTS", true),
		EnableCSP:   getBoolEnv("SECURITY_ENABLE_CSP", true),
	}

	// 許可ドメインの設定
	if domains := os.Getenv("SECURITY_ALLOWED_DOMAINS"); domains != "" {
		config.AllowedDomains = strings.Split(domains, ",")
		// 空白を削除
		for i, domain := range config.AllowedDomains {
			config.AllowedDomains[i] = strings.TrimSpace(domain)
		}
	}

	return config
}

// SecurityHeaders セキュリティヘッダーを設定するミドルウェア
func SecurityHeaders(config *SecurityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// CSP (Content Security Policy) の設定
			if config.EnableCSP {
				csp := buildCSP(config)
				w.Header().Set("Content-Security-Policy", csp)
			}

			// セキュリティヘッダーの設定
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// HSTS (HTTPS Strict Transport Security)
			if config.EnableHSTS && config.Environment == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// buildCSP CSPポリシーを構築
func buildCSP(config *SecurityConfig) string {
	var policies []string

	// デフォルトソース
	policies = append(policies, "default-src 'self'")

	// スクリプトソース
	scriptSrc := []string{
		"'self'",
		"'unsafe-inline'",     // Alpine.js等のインラインスクリプト用
		"'unsafe-eval'",       // Tailwind CDNが必要とする場合があります
		"cdn.jsdelivr.net",    // CDNライブラリ用
		"unpkg.com",           // CDNライブラリ用
		"cdn.tailwindcss.com", // Tailwind CSS CDN
	}

	// Supabase URL がある場合は追加
	if config.SupabaseURL != "" {
		scriptSrc = append(scriptSrc, config.SupabaseURL)
	}

	policies = append(policies, "script-src "+strings.Join(scriptSrc, " "))

	// スタイルソース
	styleSrc := []string{
		"'self'",
		"'unsafe-inline'", // Tailwind CSS等のインラインスタイル用
		"cdn.jsdelivr.net",
		"unpkg.com",
		"fonts.googleapis.com",
		"cdn.tailwindcss.com", // Tailwind CSS CDN
	}
	policies = append(policies, "style-src "+strings.Join(styleSrc, " "))

	// 画像ソース
	imgSrc := []string{
		"'self'",
		"data:",
		"blob:",
	}
	policies = append(policies, "img-src "+strings.Join(imgSrc, " "))

	// フォントソース
	fontSrc := []string{
		"'self'",
		"fonts.gstatic.com",
		"cdn.jsdelivr.net",
		"data:", // データURLフォント用
	}
	policies = append(policies, "font-src "+strings.Join(fontSrc, " "))

	// 接続ソース
	connectSrc := []string{
		"'self'",
	}

	// Supabase URL がある場合は追加
	if config.SupabaseURL != "" {
		connectSrc = append(connectSrc, config.SupabaseURL)
	}

	// 追加の許可ドメインがある場合は追加
	for _, domain := range config.AllowedDomains {
		if domain != "" {
			connectSrc = append(connectSrc, domain)
		}
	}

	policies = append(policies, "connect-src "+strings.Join(connectSrc, " "))

	// フレームソース
	policies = append(policies, "frame-src 'none'")

	// オブジェクトソース
	policies = append(policies, "object-src 'none'")

	// ベースURI
	policies = append(policies, "base-uri 'self'")

	// フォーム送信先
	policies = append(policies, "form-action 'self'")

	// マニフェストソース（PWA対応）
	policies = append(policies, "manifest-src 'self'")

	// ワーカーソース
	policies = append(policies, "worker-src 'self'")

	return strings.Join(policies, "; ")
}

// getEnv 環境変数を取得（デフォルト値付き）
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv 環境変数をboolで取得（デフォルト値付き）
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.ToLower(value) == "true" || value == "1"
}
