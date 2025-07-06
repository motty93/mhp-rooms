package models

import (
	"time"

	"github.com/google/uuid"
)

// PlayerName ユーザーのゲームバージョンごとのプレイヤーネーム
type PlayerName struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_player_names_user_game" json:"user_id"`
	GameVersionID uuid.UUID `gorm:"type:uuid;not null;index:idx_player_names_user_game" json:"game_version_id"`
	Name          string    `gorm:"type:varchar(50);not null" json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// リレーション
	User        User        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	GameVersion GameVersion `gorm:"foreignKey:GameVersionID;constraint:OnDelete:CASCADE" json:"game_version,omitempty"`
}

// テーブル名を明示的に指定
func (PlayerName) TableName() string {
	return "player_names"
}
