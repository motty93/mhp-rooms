package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomMessage struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoomID      uuid.UUID `gorm:"type:uuid;not null" json:"room_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Message     string    `gorm:"type:text;not null" json:"message"`
	MessageType string    `gorm:"type:varchar(20);not null;default:'chat'" json:"message_type"`
	IsDeleted   bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt   time.Time `json:"created_at"`

	// リレーション
	Room Room `gorm:"foreignKey:RoomID" json:"room"`
	User User `gorm:"foreignKey:UserID" json:"user"`
}
