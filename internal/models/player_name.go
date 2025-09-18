package models

import (
	"github.com/google/uuid"
)

// PlayerName ユーザーのゲームバージョンごとのプレイヤーネーム
type PlayerName struct {
	BaseModel
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_player_names_user_game" json:"user_id"`
	GameVersionID uuid.UUID `gorm:"type:uuid;not null;index:idx_player_names_user_game" json:"game_version_id"`
	Name          string    `gorm:"type:varchar(50);not null" json:"name"`

	// リレーション
	User        User        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	GameVersion GameVersion `gorm:"foreignKey:GameVersionID;constraint:OnDelete:CASCADE" json:"game_version,omitempty"`
}

// テーブル名を明示的に指定
func (PlayerName) TableName() string {
	return "player_names"
}
