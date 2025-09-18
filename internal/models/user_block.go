package models

import (
	"github.com/google/uuid"
)

type UserBlock struct {
	BaseModel
	BlockerUserID uuid.UUID `gorm:"type:uuid;not null" json:"blocker_user_id"`
	BlockedUserID uuid.UUID `gorm:"type:uuid;not null" json:"blocked_user_id"`
	Reason        *string   `gorm:"type:text" json:"reason"`

	// リレーション
	Blocker User `gorm:"foreignKey:BlockerUserID" json:"blocker"`
	Blocked User `gorm:"foreignKey:BlockedUserID" json:"blocked"`
}
