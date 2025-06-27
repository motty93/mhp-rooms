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

	// 接続プールの設定
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, fmt.Errorf("データベース接続プールの設定に失敗しました: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &DB{conn: conn}, nil
}

// GetConn 移行期間用
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
	// GORMのAutoMigrateが外部キー制約を自動で作成するため、
	// ここでの手動追加はコメントアウトします。
	// 必要な場合は、GORMの命名規則に従っているか確認してください。
	/*
		// 外部キー制約
		constraints := []string{
			"ALTER TABLE game_versions ADD CONSTRAINT fk_game_versions_platform FOREIGN KEY (platform_id) REFERENCES platforms(id)",
			"ALTER TABLE rooms ADD CONSTRAINT fk_rooms_game_version FOREIGN KEY (game_version_id) REFERENCES game_versions(id)",
			"ALTER TABLE rooms ADD CONSTRAINT fk_rooms_host_user FOREIGN KEY (host_user_id) REFERENCES users(id)",
			"ALTER TABLE room_members ADD CONSTRAINT fk_room_members_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE",
			"ALTER TABLE room_members ADD CONSTRAINT fk_room_members_user FOREIGN KEY (user_id) REFERENCES users(id)",
			"ALTER TABLE room_messages ADD CONSTRAINT fk_room_messages_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE",
			"ALTER TABLE room_messages ADD CONSTRAINT fk_room_messages_user FOREIGN KEY (user_id) REFERENCES users(id)",
			"ALTER TABLE user_blocks ADD CONSTRAINT fk_user_blocks_blocker FOREIGN KEY (blocker_user_id) REFERENCES users(id)",
			"ALTER TABLE user_blocks ADD CONSTRAINT fk_user_blocks_blocked FOREIGN KEY (blocked_user_id) REFERENCES users(id)",
			"ALTER TABLE room_logs ADD CONSTRAINT fk_room_logs_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE",
			"ALTER TABLE room_logs ADD CONSTRAINT fk_room_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL",
			"ALTER TABLE password_resets ADD CONSTRAINT fk_password_resets_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
			"ALTER TABLE player_names ADD CONSTRAINT fk_player_names_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
			"ALTER TABLE player_names ADD CONSTRAINT fk_player_names_game_version FOREIGN KEY (game_version_id) REFERENCES game_versions(id) ON DELETE CASCADE",
		}

		// チェック制約
		checks := []string{
			"ALTER TABLE users ADD CONSTRAINT chk_users_supabase_user_id CHECK (supabase_user_id IS NOT NULL)",
			"ALTER TABLE rooms ADD CONSTRAINT chk_rooms_max_players CHECK (max_players = 4)",
		}

		// ユニーク制約
		uniques := []string{
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_room_members_active ON room_members(room_id, user_id) WHERE status = 'active'",
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_blocks_unique ON user_blocks(blocker_user_id, blocked_user_id)",
		}

		// 一般的なインデックス
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_users_is_active_created_at ON users(is_active, created_at)",
			"CREATE INDEX IF NOT EXISTS idx_game_versions_is_active_display_order ON game_versions(is_active, display_order)",
			"CREATE INDEX IF NOT EXISTS idx_rooms_game_version_status_is_active ON rooms(game_version_id, status, is_active)",
			"CREATE INDEX IF NOT EXISTS idx_rooms_host_user_id ON rooms(host_user_id)",
			"CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at DESC)",
			"CREATE INDEX IF NOT EXISTS idx_room_members_user_id_status ON room_members(user_id, status)",
			"CREATE INDEX IF NOT EXISTS idx_room_members_room_id_player_number ON room_members(room_id, player_number)",
			"CREATE INDEX IF NOT EXISTS idx_room_messages_room_id_created_at ON room_messages(room_id, created_at DESC)",
			"CREATE INDEX IF NOT EXISTS idx_room_messages_user_id ON room_messages(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_user_blocks_blocked_user_id ON user_blocks(blocked_user_id)",
			"CREATE INDEX IF NOT EXISTS idx_room_logs_room_id_created_at ON room_logs(room_id, created_at DESC)",
			"CREATE INDEX IF NOT EXISTS idx_room_logs_user_id ON room_logs(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_room_logs_action ON room_logs(action)",
			"CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token)",
			"CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at)",
			"CREATE INDEX IF NOT EXISTS idx_player_names_user_game ON player_names(user_id, game_version_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uk_player_names_user_game ON player_names(user_id, game_version_id)",
		}

		// すべてのSQL文を実行
		allStatements := append(append(constraints, checks...), append(uniques, indexes...)...)

		for _, stmt := range allStatements {
			if err := db.conn.Exec(stmt).Error; err != nil {
				// 制約やインデックスの失敗は警告レベルで継続
			}
		}
	*/
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
