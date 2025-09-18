package persistence

import (
	"fmt"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/infrastructure/persistence/turso"
)

// NewDBAdapter はデータベースタイプに応じて適切なアダプターを作成
func NewDBAdapter(cfg *config.Config) (DBAdapter, error) {
	switch cfg.Database.Type {
	case "turso":
		return turso.NewDB(cfg)
	case "postgres":
		return postgres.NewDB(cfg)
	default:
		return nil, fmt.Errorf("不明なデータベースタイプ: %s", cfg.Database.Type)
	}
}
