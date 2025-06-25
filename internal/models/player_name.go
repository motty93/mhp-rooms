package models

import (
	"time"

	"github.com/google/uuid"
)

// PlayerName ユーザーのゲームバージョンごとのプレイヤーネーム
type PlayerName struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	GameVersion string    `gorm:"type:varchar(10);not null" json:"game_version"` // MHP, MHP2, MHP2G, MHP3
	PlayerName  string    `gorm:"type:varchar(100);not null" json:"player_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// リレーション
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// テーブル名を明示的に指定
func (PlayerName) TableName() string {
	return "player_names"
}