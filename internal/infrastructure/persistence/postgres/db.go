package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/models"
)

type DB struct {
	conn *gorm.DB
}

func NewDB(cfg *config.Config) (*DB, error) {
	dsn := cfg.GetDSN()

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("データベース接続に失敗しました: %w", err)
	}

	sqlDB, err := conn.DB()
	if err != nil {
		return nil, fmt.Errorf("データベース接続プールの設定に失敗しました: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &DB{conn: conn}, nil
}

func (db *DB) GetConn() *gorm.DB {
	return db.conn
}

func (db *DB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) Migrate() error {
	// テーブル作成順序に注意（外部キー制約のため）
	err := db.conn.AutoMigrate(
		&models.Platform{},
		&models.User{},
		&models.GameVersion{},
		&models.PlayerName{},
		&models.Room{},
		&models.RoomMember{},
		&models.RoomMessage{},
		&models.UserBlock{},
		&models.RoomLog{},
		&models.PasswordReset{},
	)
	if err != nil {
		return fmt.Errorf("マイグレーションに失敗しました: %w", err)
	}

	// 制約とインデックスを追加
	if err := db.addConstraintsAndIndexes(); err != nil {
		return fmt.Errorf("制約とインデックスの追加に失敗しました: %w", err)
	}

	// 初期データを挿入
	if err := db.insertInitialData(); err != nil {
		return fmt.Errorf("初期データの挿入に失敗しました: %w", err)
	}

	return nil
}

func (db *DB) addConstraintsAndIndexes() error {
	// ただし、チェック制約、ユニーク制約、パフォーマンス用のインデックスは手動で追加
	// チェック制約
	checks := []string{
		"ALTER TABLE rooms ADD CONSTRAINT IF NOT EXISTS chk_rooms_max_players CHECK (max_players = 4)",
	}

	// ユニーク制約
	uniques := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_room_members_active ON room_members(room_id, user_id) WHERE status = 'active'",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_blocks_unique ON user_blocks(blocker_user_id, blocked_user_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS uk_player_names_user_game ON player_names(user_id, game_version_id)",
	}

	// パフォーマンス用インデックス
	indexes := []string{
		// ユーザー関連
		"CREATE INDEX IF NOT EXISTS idx_users_is_active_created_at ON users(is_active, created_at)",

		// ゲームバージョン関連
		"CREATE INDEX IF NOT EXISTS idx_game_versions_is_active_display_order ON game_versions(is_active, display_order)",

		// ルーム関連
		"CREATE INDEX IF NOT EXISTS idx_rooms_game_version_status_is_active ON rooms(game_version_id, status, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_host_user_id ON rooms(host_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at DESC)",

		// ルームメンバー関連
		"CREATE INDEX IF NOT EXISTS idx_room_members_user_id_status ON room_members(user_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_room_members_room_id_player_number ON room_members(room_id, player_number)",

		// ルームメッセージ関連
		"CREATE INDEX IF NOT EXISTS idx_room_messages_room_id_created_at ON room_messages(room_id, created_at DESC)",

		// ユーザーブロック関連
		"CREATE INDEX IF NOT EXISTS idx_user_blocks_blocked_user_id ON user_blocks(blocked_user_id)",

		// ルームログ関連
		"CREATE INDEX IF NOT EXISTS idx_room_logs_room_id_created_at ON room_logs(room_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_room_logs_action ON room_logs(action)",

		// パスワードリセット関連
		"CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token)",
		"CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at)",
	}

	// すべてのSQL文を実行
	allStatements := append(append(checks, uniques...), indexes...)

	for _, stmt := range allStatements {
		if err := db.conn.Exec(stmt).Error; err != nil {
			// 既に存在する制約やインデックスの場合はエラーを無視
			// IF NOT EXISTSを使用しているため、通常はエラーは発生しない
		}
	}

	return nil
}

func (db *DB) insertInitialData() error {
	tx := db.conn.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Platforms
	var platformCount int64
	tx.Model(&models.Platform{}).Count(&platformCount)
	var playstationPlatform models.Platform

	if platformCount == 0 {
		playstationPlatform = models.Platform{
			Name:         "PlayStation",
			DisplayOrder: 1,
		}
		if err := tx.Create(&playstationPlatform).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("プラットフォームの挿入に失敗しました: %w", err)
		}
	} else {
		tx.First(&playstationPlatform, "name = ?", "PlayStation")
	}

	// GameVersions
	var gameVersionCount int64
	tx.Model(&models.GameVersion{}).Count(&gameVersionCount)
	if gameVersionCount == 0 {
		gameVersions := []models.GameVersion{
			{
				Code:         "MHP",
				Name:         "モンスターハンターポータブル",
				DisplayOrder: 1,
				IsActive:     true,
				PlatformID:   playstationPlatform.ID,
			},
			{
				Code:         "MHP2",
				Name:         "モンスターハンターポータブル 2nd",
				DisplayOrder: 2,
				IsActive:     true,
				PlatformID:   playstationPlatform.ID,
			},
			{
				Code:         "MHP2G",
				Name:         "モンスターハンターポータブル 2ndG",
				DisplayOrder: 3,
				IsActive:     true,
				PlatformID:   playstationPlatform.ID,
			},
			{
				Code:         "MHP3",
				Name:         "モンスターハンターポータブル 3rd",
				DisplayOrder: 4,
				IsActive:     true,
				PlatformID:   playstationPlatform.ID,
			},
		}

		for _, gv := range gameVersions {
			if err := tx.Create(&gv).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("ゲームバージョンの挿入に失敗しました: %w", err)
			}
		}
	}

	return tx.Commit().Error
}
