package models

import (
	"time"

	"github.com/google/uuid"
)

type MessageReaction struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MessageID    uuid.UUID `gorm:"type:uuid;not null" json:"message_id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ReactionType string    `gorm:"type:varchar(50);not null" json:"reaction_type"`
	CreatedAt    time.Time `json:"created_at"`

	// リレーション
	Message RoomMessage `gorm:"foreignKey:MessageID" json:"-"`
	User    User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName はGORMで使用するテーブル名を指定
func (MessageReaction) TableName() string {
	return "message_reactions"
}

// MessageReactionCount はリアクションの集計結果を表す構造体
type MessageReactionCount struct {
	MessageID     uuid.UUID   `json:"message_id"`
	ReactionType  string      `json:"reaction_type"`
	Emoji         string      `json:"emoji"`
	ReactionName  string      `json:"reaction_name"`
	ReactionCount int         `json:"reaction_count"`
	UserIDs       []uuid.UUID `json:"user_ids,omitempty"`
	HasReacted    bool        `json:"has_reacted,omitempty"` // 現在のユーザーがリアクションしているか
}

// UserReactionState はユーザーのリアクション状態を表す構造体
type UserReactionState struct {
	MessageID     uuid.UUID `json:"message_id"`
	ReactionTypes []string  `json:"reaction_types"`
}
