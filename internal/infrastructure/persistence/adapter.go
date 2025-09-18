package persistence

import (
	"gorm.io/gorm"
	"mhp-rooms/internal/models"
)

// DBAdapter はデータベースアダプターのインターフェース
type DBAdapter interface {
	// GetConn はGORMのデータベース接続を返す
	GetConn() *gorm.DB

	// Close はデータベース接続を閉じる
	Close() error

	// Migrate はデータベースマイグレーションを実行する
	Migrate() error

	// GetType はデータベースの種類を返す（"sqlite" or "postgres"）
	GetType() string
}

// MigrationHelper はデータベース固有のマイグレーション処理を提供
type MigrationHelper interface {
	// CreateConstraintsAndIndexes はDB固有の制約とインデックスを作成
	CreateConstraintsAndIndexes(db *gorm.DB) error

	// InsertInitialData は初期データを挿入
	InsertInitialData(db *gorm.DB) error
}

// CommonMigrate はデータベース共通のマイグレーションを実行
func CommonMigrate(db *gorm.DB) error {
	// 共通のテーブル作成
	return db.AutoMigrate(
		&models.Platform{},
		&models.GameVersion{},
		&models.User{},
		&models.Room{},
		&models.RoomMember{},
		&models.RoomMessage{},
		&models.MessageReaction{},
		&models.ReactionType{},
		&models.UserBlock{},
		&models.PlayerName{},
		&models.UserFollow{},
		&models.UserActivity{},
		&models.RoomLog{},
		&models.PasswordReset{},
	)
}
