package models

import (
	"time"

	"github.com/google/uuid"
)

// RoomLog はルームアクションの監査ログ
type RoomLog struct {
	ID        uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoomID    uuid.UUID              `gorm:"type:uuid;not null" json:"room_id"`
	UserID    *uuid.UUID             `gorm:"type:uuid" json:"user_id"`
	Action    string                 `gorm:"type:varchar(50);not null" json:"action"`
	Details   map[string]interface{} `gorm:"type:jsonb" json:"details"`
	CreatedAt time.Time              `json:"created_at"`

	// リレーション
	Room Room  `gorm:"foreignKey:RoomID" json:"room"`
	User *User `gorm:"foreignKey:UserID" json:"user"`
}
