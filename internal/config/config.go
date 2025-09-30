package config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Database    DatabaseConfig
	Server      ServerConfig
	Environment string
	ServiceMode string // "main" | "sse" | "both"
	Migration   MigrationConfig
	Debug       DebugConfig
	GCS         GCSConfig
	Redis       RedisConfig
	SSE         SSEConfig
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
	PrivateBucket  string // 通報用プライベートバケット
	MaxUploadBytes int64
	AllowedMIMEs   map[string]struct{}
	AssetPrefix    string
}

type RedisConfig struct {
	URL            string
	Enabled        bool
	Host           string
	Port           string
	Password       string
	DB             int
	MaxRetries     int
	RetryInterval  time.Duration
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

type SSEConfig struct {
	TokenTTL          time.Duration
	Host              string // SSEコンテナの公開URL
	PingInterval      time.Duration
	ConnectionTimeout time.Duration
}

var AppConfig *Config

func Init() {
	AppConfig = &Config{
		Database: DatabaseConfig{
			Type:           GetEnv("DB_TYPE", "turso"),
			URL:            GetEnv("DATABASE_URL", ""),
			Host:           GetEnv("DB_HOST", "localhost"),
			Port:           GetEnv("DB_PORT", "5432"),
			User:           GetEnv("DB_USER", "postgres"),
			Password:       GetEnv("DB_PASSWORD", "postgres"),
			Name:           GetEnv("DB_NAME", "mhp_rooms"),
			SSLMode:        GetEnv("DB_SSLMODE", getDefaultSSLMode()),
			TursoURL:       GetEnv("TURSO_DATABASE_URL", ""),
			TursoAuthToken: GetEnv("TURSO_AUTH_TOKEN", ""),
		},
		Server: ServerConfig{
			Port:    GetEnv("PORT", "8080"),
			Host:    GetEnv("HOST", "0.0.0.0"),
			SSEHost: GetEnv("SSE_HOST", ""), // 空の場合は同一サーバー
		},
		Environment: GetEnv("ENV", "development"),
		ServiceMode: GetEnv("SERVICE_MODE", "main"), // "main" または "sse"
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
			PrivateBucket:  MustGetEnv("GCS_PRIVATE_BUCKET"),        // 通報用プライベートバケット
			MaxUploadBytes: GetEnvInt64("MAX_UPLOAD_BYTES", 10<<20), // デフォルト10MB
			AllowedMIMEs:   parseAllowedMIMEs(GetEnv("ALLOW_CONTENT_TYPES", ""), []string{"image/jpeg", "image/png", "image/webp"}),
			AssetPrefix:    cleanAssetPrefix(GetEnv("ASSET_PREFIX", "")),
		},
		Redis: RedisConfig{
			URL:            os.Getenv("REDIS_URL"),
			Enabled:        getEnvBool("REDIS_ENABLED", false),
			Host:           GetEnv("REDIS_HOST", "localhost"),
			Port:           GetEnv("REDIS_PORT", "6379"),
			Password:       os.Getenv("REDIS_PASSWORD"),
			DB:             getEnvInt("REDIS_DB", 0),
			MaxRetries:     getEnvInt("REDIS_MAX_RETRIES", 3),
			RetryInterval:  getEnvDuration("REDIS_RETRY_INTERVAL", 1*time.Second),
			ConnectTimeout: getEnvDuration("REDIS_CONNECT_TIMEOUT", 5*time.Second),
			ReadTimeout:    getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:   getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},

		SSE: SSEConfig{
			TokenTTL:          getEnvDuration("SSE_TOKEN_TTL", 5*time.Minute),
			Host:              os.Getenv("SSE_HOST"),
			PingInterval:      getEnvDuration("SSE_PING_INTERVAL", 30*time.Second),
			ConnectionTimeout: getEnvDuration("SSE_CONNECTION_TIMEOUT", 30*time.Minute),
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

func GetEnv(key, defaultValue string) string {
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
	env := GetEnv("ENV", "development")
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

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	result, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("Warning: invalid duration value for %s: %s", key, value)
		return defaultValue
	}

	return result
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
