package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database    DatabaseConfig
	Server      ServerConfig
	Environment string
	Migration   MigrationConfig
	Debug       DebugConfig
}

type DebugConfig struct {
	AuthLogs bool
	SQLLogs  bool
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServerConfig struct {
	Port string
	Host string
}

type MigrationConfig struct {
	AutoRun bool
}

var AppConfig *Config

func Init() {
	AppConfig = &Config{
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", ""),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "mhp_rooms"),
			SSLMode:  getEnv("DB_SSLMODE", getDefaultSSLMode()),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		Environment: getEnv("ENV", "development"),
		Migration: MigrationConfig{
			AutoRun: getEnvBool("RUN_MIGRATION", false),
		},
		Debug: DebugConfig{
			AuthLogs: getEnvBool("DEBUG_AUTH_LOGS", false),
			SQLLogs:  getEnvBool("DEBUG_SQL_LOGS", false),
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
