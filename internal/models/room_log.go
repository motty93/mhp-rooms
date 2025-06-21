package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoomLog はルームログを管理するモデル
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

// BeforeCreate はレコード作成前にUUIDを生成
func (rl *RoomLog) BeforeCreate(tx *gorm.DB) error {
	if rl.ID == uuid.Nil {
		rl.ID = uuid.New()
	}
	return nil
}