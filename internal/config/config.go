package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Database    DatabaseConfig
	Server      ServerConfig
	Environment string
	ServiceMode string // "main" または "sse"
	Migration   MigrationConfig
	Debug       DebugConfig
	GCS         GCSConfig
}

type DebugConfig struct {
	AuthLogs bool
	SQLLogs  bool
}

type DatabaseConfig struct {
	Type           string // "turso" or "postgres"
	URL            string
	Host           string
	Port           string
	User           string
	Password       string
	Name           string
	SSLMode        string
	TursoURL       string // Turso用のデータベースURL
	TursoAuthToken string // Turso用の認証トークン
}

type ServerConfig struct {
	Port    string
	Host    string
	SSEHost string // SSEサーバーのホスト（空の場合は同一サーバー）
}

type MigrationConfig struct {
	AutoRun bool
}

type GCSConfig struct {
	Bucket         string
	BaseURL        string
	MaxUploadBytes int64
	AllowedMIMEs   map[string]struct{}
	AssetPrefix    string
}

var AppConfig *Config

func Init() {
	AppConfig = &Config{
		Database: DatabaseConfig{
			Type:           getEnv("DB_TYPE", "turso"),
			URL:            getEnv("DATABASE_URL", ""),
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnv("DB_PORT", "5432"),
			User:           getEnv("DB_USER", "postgres"),
			Password:       getEnv("DB_PASSWORD", "postgres"),
			Name:           getEnv("DB_NAME", "mhp_rooms"),
			SSLMode:        getEnv("DB_SSLMODE", getDefaultSSLMode()),
			TursoURL:       getEnv("TURSO_DATABASE_URL", ""),
			TursoAuthToken: getEnv("TURSO_AUTH_TOKEN", ""),
		},
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			Host:    getEnv("HOST", "0.0.0.0"),
			SSEHost: getEnv("SSE_HOST", ""), // 空の場合は同一サーバー
		},
		Environment: getEnv("ENV", "development"),
		ServiceMode: getEnv("SERVICE_MODE", "main"), // "main" または "sse"
		Migration: MigrationConfig{
			AutoRun: getEnvBool("RUN_MIGRATION", false),
		},
		Debug: DebugConfig{
			AuthLogs: getEnvBool("DEBUG_AUTH_LOGS", false),
			SQLLogs:  getEnvBool("DEBUG_SQL_LOGS", false),
		},
		GCS: GCSConfig{
			Bucket:         MustGetEnv("GCS_BUCKET"),
			BaseURL:        MustGetEnv("BASE_PUBLIC_ASSET_URL"),
			MaxUploadBytes: GetEnvInt64("MAX_UPLOAD_BYTES", 10<<20), // デフォルト10MB
			AllowedMIMEs:   parseAllowedMIMEs(getEnv("ALLOW_CONTENT_TYPES", ""), []string{"image/jpeg", "image/png", "image/webp"}),
			AssetPrefix:    cleanAssetPrefix(getEnv("ASSET_PREFIX", "")),
		},
	}
}

// データベース接続文字列を取得
func (c *Config) GetDSN() string {
	// DATABASE_URLが設定されている場合は優先的に使用
	if c.Database.URL != "" {
		return c.Database.URL
	}

	// 個別の環境変数から構築
	return "host=" + c.Database.Host +
		" port=" + c.Database.Port +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.Name +
		" sslmode=" + c.Database.SSLMode
}

// IsProduction 本番環境かどうかを判定
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return !c.IsProduction()
}

func (c *Config) GetServerAddr() string {
	return c.Server.Host + ":" + c.Server.Port
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDefaultSSLMode() string {
	env := getEnv("ENV", "development")
	if env == "production" {
		return "require"
	}
	return "disable"
}

// MustGetEnv 必須環境変数の取得
func MustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("環境変数 %s が設定されていません", key))
	}
	return v
}

// GetEnvInt64 int64型環境変数の取得
func GetEnvInt64(key string, defaultValue int64) int64 {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("環境変数 %s が無効な数値です: %v", key, err))
	}
	return v
}

// parseAllowedMIMEs 許可されたMIMEタイプをパース
func parseAllowedMIMEs(env string, defaults []string) map[string]struct{} {
	list := defaults
	if env != "" {
		list = strings.Split(env, ",")
	}
	m := make(map[string]struct{})
	for _, s := range list {
		m[strings.TrimSpace(s)] = struct{}{}
	}
	return m
}

// cleanAssetPrefix プレフィックスをクリーンアップ（dev/stg/prodなどのみ許可）
func cleanAssetPrefix(s string) string {
	s = strings.TrimSpace(strings.Trim(s, "/"))
	if s == "" {
		return "dev" // デフォルトをdevに
	}
	re := regexp.MustCompile(`^[a-z0-9._-]+$`)
	if !re.MatchString(s) {
		panic(fmt.Sprintf("無効なASSET_PREFIX: %s", s))
	}
	return s
}
