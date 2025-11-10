package middleware

import (
	"net/http"
	"os"
	"strings"
)

type SecurityConfig struct {
	SupabaseURL    string
	Environment    string
	AllowedDomains []string
	EnableHSTS     bool
	EnableCSP      bool
}

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
		for i, domain := range config.AllowedDomains {
			config.AllowedDomains[i] = strings.TrimSpace(domain)
		}
	}

	return config
}

// セキュリティヘッダーを設定するミドルウェア
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
		"https://www.googletagmanager.com", // Google Tag Manager/Analytics
		"https://www.google-analytics.com", // Google Analytics
		"https://pagead2.googlesyndication.com", // Google AdSense
		"https://adservice.google.com", // Google Ad Service
		"https://googleads.g.doubleclick.net", // Google DoubleClick
	}

	// Supabase URL がある場合は追加（末尾スラッシュを削除して正規化）
	if config.SupabaseURL != "" {
		supabaseURL := strings.TrimSuffix(config.SupabaseURL, "/")
		// http://を強制的にhttps://に変換
		if strings.HasPrefix(supabaseURL, "http://") {
			supabaseURL = "https://" + strings.TrimPrefix(supabaseURL, "http://")
		}
		scriptSrc = append(scriptSrc, supabaseURL)
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
		"https://pagead2.googlesyndication.com", // AdSense広告画像
		"https://googleads.g.doubleclick.net", // DoubleClick広告画像
		"https://tpc.googlesyndication.com", // AdSense tracking pixels
	}

	// GCS/CDN URLを追加
	if baseAssetURL := os.Getenv("BASE_PUBLIC_ASSET_URL"); baseAssetURL != "" {
		// URLからドメイン部分を抽出
		if strings.HasPrefix(baseAssetURL, "https://") {
			domain := strings.TrimPrefix(baseAssetURL, "https://")
			// パスが含まれる場合は削除
			if idx := strings.Index(domain, "/"); idx > 0 {
				domain = domain[:idx]
			}
			imgSrc = append(imgSrc, "https://"+domain)
		} else if strings.HasPrefix(baseAssetURL, "http://") {
			domain := strings.TrimPrefix(baseAssetURL, "http://")
			// パスが含まれる場合は削除
			if idx := strings.Index(domain, "/"); idx > 0 {
				domain = domain[:idx]
			}
			imgSrc = append(imgSrc, "http://"+domain)
		}
	}

	// Supabase URL がある場合は画像ソースにも追加（ストレージ用）
	if config.SupabaseURL != "" {
		supabaseURL := strings.TrimSuffix(config.SupabaseURL, "/")
		if strings.HasPrefix(supabaseURL, "http://") {
			supabaseURL = "https://" + strings.TrimPrefix(supabaseURL, "http://")
		}
		imgSrc = append(imgSrc, supabaseURL)
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
		"https://www.google-analytics.com", // Google Analyticsデータ送信用
		"https://pagead2.googlesyndication.com", // AdSense通信用
		"https://googleads.g.doubleclick.net", // DoubleClick通信用
	}

	// Supabase URL がある場合は追加（API通信用）
	if config.SupabaseURL != "" {
		supabaseURL := strings.TrimSuffix(config.SupabaseURL, "/")
		if strings.HasPrefix(supabaseURL, "http://") {
			supabaseURL = "https://" + strings.TrimPrefix(supabaseURL, "http://")
		}
		connectSrc = append(connectSrc, supabaseURL)
	}

	// 追加の許可ドメインがある場合は追加
	for _, domain := range config.AllowedDomains {
		if domain != "" {
			connectSrc = append(connectSrc, domain)
		}
	}

	policies = append(policies, "connect-src "+strings.Join(connectSrc, " "))

	// フレームソース（AdSenseは iframe を使用）
	frameSrc := []string{
		"https://googleads.g.doubleclick.net",
		"https://tpc.googlesyndication.com",
		"https://www.google.com", // reCAPTCHA等
	}
	policies = append(policies, "frame-src "+strings.Join(frameSrc, " "))

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
