package turso

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
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

	// プラットフォームの初期データ
	platforms := []models.Platform{
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Name:         "Playstation Portable",
			DisplayOrder: 1,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Name:         "Nintendo",
			DisplayOrder: 2,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Name:         "Playstation",
			DisplayOrder: 3,
		},
	}

	for _, platform := range platforms {
		if err := tx.FirstOrCreate(&platform, models.Platform{BaseModel: models.BaseModel{ID: platform.ID}}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// ゲームバージョンの初期データ
	gameVersions := []models.GameVersion{
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Code:         "MHP",
			Name:         "モンスターハンターポータブル",
			PlatformID:   platforms[0].ID,
			DisplayOrder: 1,
			IsActive:     true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Code:         "MHP2",
			Name:         "モンスターハンターポータブル 2nd",
			PlatformID:   platforms[0].ID,
			DisplayOrder: 2,
			IsActive:     true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Code:         "MHP2G",
			Name:         "モンスターハンターポータブル 2nd G",
			PlatformID:   platforms[0].ID,
			DisplayOrder: 3,
			IsActive:     true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Code:         "MHP3",
			Name:         "モンスターハンターポータブル 3rd",
			PlatformID:   platforms[0].ID,
			DisplayOrder: 4,
			IsActive:     true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Code:         "MHXX",
			Name:         "モンスターハンターダブルクロス",
			PlatformID:   platforms[1].ID,
			DisplayOrder: 5,
			IsActive:     true,
		},
	}

	for _, gameVersion := range gameVersions {
		if err := tx.FirstOrCreate(&gameVersion, models.GameVersion{BaseModel: models.BaseModel{ID: gameVersion.ID}}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// リアクションタイプの初期データ
	reactionTypes := []models.ReactionType{
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Code:         "like",
			Name:         "いいね",
			Emoji:        "👍",
			DisplayOrder: 1,
			IsActive:     true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Code:         "heart",
			Name:         "ハート",
			Emoji:        "❤ ",
			DisplayOrder: 2,
			IsActive:     true,
		},
	}

	for _, reactionType := range reactionTypes {
		if err := tx.FirstOrCreate(&reactionType, models.ReactionType{BaseModel: models.BaseModel{ID: reactionType.ID}}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// commonMigrate はデータベース共通のマイグレーションを実行
func (db *DB) commonMigrate() error {
	// 共通のテーブル作成
	return db.conn.AutoMigrate(
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
