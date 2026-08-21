package turso

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/models"
)

type DB struct {
	conn *gorm.DB
}

// NewDB はTursoデータベース接続を作成
func NewDB(cfg *config.Config) (*DB, error) {
	if cfg.Database.TursoURL == "" {
		return nil, fmt.Errorf("TURSO_DATABASE_URL が設定されていません")
	}
	if cfg.Database.TursoAuthToken == "" {
		return nil, fmt.Errorf("TURSO_AUTH_TOKEN が設定されていません")
	}

	logLevel := logger.Warn
	if cfg.Debug.SQLLogs {
		logLevel = logger.Info
	}

	// Turso接続設定
	dsn := fmt.Sprintf("%s?authToken=%s", cfg.Database.TursoURL, cfg.Database.TursoAuthToken)
	sqlDB, err := sql.Open("libsql", dsn)
	if err != nil {
		return nil, fmt.Errorf("Tursoデータベース接続の作成に失敗しました: %w", err)
	}

	// 接続テスト
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("Tursoデータベース接続に失敗しました: %w", err)
	}

	// GORM設定
	conn, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("GORM初期化に失敗しました: %w", err)
	}

	// 接続プール設定
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &DB{conn: conn}, nil
}

func (db *DB) GetConn() *gorm.DB {
	return db.conn
}

func (db *DB) GetType() string {
	return "turso"
}

func (db *DB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) Migrate() error {
	// 共通のマイグレーションを実行
	if err := db.commonMigrate(); err != nil {
		return fmt.Errorf("マイグレーションに失敗しました: %w", err)
	}

	// Turso固有の制約とインデックスを追加
	if err := db.createConstraintsAndIndexes(); err != nil {
		return fmt.Errorf("制約とインデックスの追加に失敗しました: %w", err)
	}

	// 初期データを挿入
	if err := db.insertInitialData(); err != nil {
		return fmt.Errorf("初期データの挿入に失敗しました: %w", err)
	}

	return nil
}

func (db *DB) createConstraintsAndIndexes() error {
	// Turso用のインデックスとチェック制約
	indexes := []string{
		// ユニーク制約
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_room_members_active ON room_members(room_id, user_id) WHERE status = 'active'",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_blocks_unique ON user_blocks(blocker_user_id, blocked_user_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS uk_player_names_user_game ON player_names(user_id, game_version_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_follows_unique ON user_follows(follower_user_id, following_user_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_message_reactions_unique ON message_reactions(message_id, user_id, reaction_type)",

		// パフォーマンス用インデックス
		"CREATE INDEX IF NOT EXISTS idx_users_is_active_created_at ON users(is_active, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_users_supabase_user_id ON users(supabase_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_game_versions_is_active_display_order ON game_versions(is_active, display_order)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_game_version_is_active ON rooms(game_version_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_host_user_id ON rooms(host_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_room_members_user_id_status ON room_members(user_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_room_members_room_id_player_number ON room_members(room_id, player_number)",
		"CREATE INDEX IF NOT EXISTS idx_room_messages_room_id_created_at ON room_messages(room_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_id_created_at ON notifications(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_user_activities_activity_type_created_at ON user_activities(activity_type, created_at DESC)",
	}

	for _, stmt := range indexes {
		if err := db.conn.Exec(stmt).Error; err != nil {
			// インデックスが既に存在する場合のエラーは無視
			fmt.Printf("インデックス作成時の警告: %v\n", err)
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

	// プラットフォームの初期データを挿入（既に存在する場合はスキップ）
	var platformCount int64
	tx.Model(&models.Platform{}).Count(&platformCount)

	if platformCount == 0 {
		platforms := []models.Platform{
			{Name: "Playstation Portable", DisplayOrder: 1},
			{Name: "Nintendo", DisplayOrder: 2},
			{Name: "Playstation", DisplayOrder: 3},
		}

		for _, platform := range platforms {
			if err := tx.Create(&platform).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("プラットフォームの挿入に失敗しました: %w", err)
			}
		}
	}

	// ゲームバージョンの初期データを挿入（既に存在する場合はスキップ）
	var gameVersionCount int64
	tx.Model(&models.GameVersion{}).Count(&gameVersionCount)

	if gameVersionCount == 0 {
		// プラットフォームを取得
		var pspPlatform, nintendoPlatform models.Platform
		tx.First(&pspPlatform, "name = ?", "Playstation Portable")
		tx.First(&nintendoPlatform, "name = ?", "Nintendo")

		gameVersions := []models.GameVersion{
			{Code: "MHP", Name: "モンスターハンターポータブル", PlatformID: pspPlatform.ID, DisplayOrder: 1, IsActive: true},
			{Code: "MHP2", Name: "モンスターハンターポータブル 2nd", PlatformID: pspPlatform.ID, DisplayOrder: 2, IsActive: true},
			{Code: "MHP2G", Name: "モンスターハンターポータブル 2nd G", PlatformID: pspPlatform.ID, DisplayOrder: 3, IsActive: true},
			{Code: "MHP3", Name: "モンスターハンターポータブル 3rd", PlatformID: pspPlatform.ID, DisplayOrder: 4, IsActive: true},
			{Code: "MHXX", Name: "モンスターハンターダブルクロス", PlatformID: nintendoPlatform.ID, DisplayOrder: 5, IsActive: true},
		}

		for _, gameVersion := range gameVersions {
			if err := tx.Create(&gameVersion).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("ゲームバージョンの挿入に失敗しました: %w", err)
			}
		}
	}

	// リアクションタイプの初期データを挿入（既に存在する場合はスキップ）
	var reactionTypeCount int64
	tx.Model(&models.ReactionType{}).Count(&reactionTypeCount)

	if reactionTypeCount == 0 {
		reactionTypes := []models.ReactionType{
			{Code: "like", Name: "いいね", Emoji: "👍", DisplayOrder: 1, IsActive: true},
			{Code: "heart", Name: "ハート", Emoji: "❤ ", DisplayOrder: 2, IsActive: true},
		}

		for _, reactionType := range reactionTypes {
			if err := tx.Create(&reactionType).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("リアクションタイプの挿入に失敗しました: %w", err)
			}
		}
	}

	return tx.Commit().Error
}

// commonMigrate はデータベース共通のマイグレーションを実行
func (db *DB) commonMigrate() error {
	// 共通のテーブル作成（モデル一覧は models.AllModels で一元管理）
	return db.conn.AutoMigrate(models.AllModels()...)
}
