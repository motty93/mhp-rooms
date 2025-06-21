package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserBlock struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BlockerUserID uuid.UUID `gorm:"type:uuid;not null" json:"blocker_user_id"`
	BlockedUserID uuid.UUID `gorm:"type:uuid;not null" json:"blocked_user_id"`
	Reason        *string   `gorm:"type:text" json:"reason"`
	CreatedAt     time.Time `json:"created_at"`

	// リレーション
	Blocker User `gorm:"foreignKey:BlockerUserID" json:"blocker"`
	Blocked User `gorm:"foreignKey:BlockedUserID" json:"blocked"`
}

func (ub *UserBlock) BeforeCreate(tx *gorm.DB) error {
	if ub.ID == uuid.Nil {
		ub.ID = uuid.New()
	}
	return nil
}