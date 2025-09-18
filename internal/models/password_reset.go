package models

import (
	"time"

	"github.com/google/uuid"
)

type PasswordReset struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"type:varchar(255);not null;unique;index" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"not null;default:false" json:"used"`

	// リレーション
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// IsExpired トークンが期限切れかどうかを確認
func (pr *PasswordReset) IsExpired() bool {
	return time.Now().After(pr.ExpiresAt)
}

// IsValid トークンが有効かどうかを確認
func (pr *PasswordReset) IsValid() bool {
	return !pr.Used && !pr.IsExpired()
}
