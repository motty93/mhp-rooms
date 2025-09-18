package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomMember struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	RoomID       uuid.UUID  `gorm:"type:uuid;not null" json:"room_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	PlayerNumber int        `gorm:"not null" json:"player_number"`
	IsHost       bool       `gorm:"not null;default:false" json:"is_host"`
	Status       string     `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	JoinedAt     time.Time  `json:"joined_at"`
	LeftAt       *time.Time `json:"left_at"`

	// リレーション
	Room Room `gorm:"foreignKey:RoomID" json:"room"`
	User User `gorm:"foreignKey:UserID" json:"user"`

	// 表示用フィールド（DBには保存されない）
	DisplayName string `gorm:"-" json:"display_name,omitempty"`
}
