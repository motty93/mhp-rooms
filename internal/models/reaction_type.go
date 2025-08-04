package models

import (
	"time"

	"github.com/google/uuid"
)

type ReactionType struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code         string    `gorm:"type:varchar(50);unique;not null" json:"code"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	Emoji        string    `gorm:"type:varchar(10);not null" json:"emoji"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName はGORMで使用するテーブル名を指定
func (ReactionType) TableName() string {
	return "reaction_types"
}

// GetActiveReactionTypes はアクティブなリアクションタイプを取得するための構造体
type ActiveReactionTypes struct {
	Types []ReactionType `json:"types"`
}
