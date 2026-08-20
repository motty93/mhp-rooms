package models

import (
	"time"

	"github.com/google/uuid"
)

// 通知の種類（notifications.type）
const (
	NotificationRoomAutoDismissed = "room_auto_dismissed" // 作成した部屋が一定期間活動がなく自動削除された
	NotificationRoomKicked        = "room_kicked"         // 部屋からホストにより退出させられた
	NotificationRoomDismissed     = "room_dismissed"      // 参加していた部屋がホストにより解散された
	NotificationFollow            = "follow"              // フォローされた
)

// Notification ユーザー宛のお知らせ
type Notification struct {
	BaseModel
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Type        string     `gorm:"type:varchar(30);not null" json:"type"`
	Title       string     `gorm:"type:varchar(200);not null" json:"title"`
	Body        *string    `gorm:"type:text" json:"body"`
	LinkURL     *string    `gorm:"type:varchar(500)" json:"link_url"`
	ActorUserID *uuid.UUID `gorm:"type:uuid" json:"actor_user_id"`
	ReadAt      *time.Time `json:"read_at"`

	// リレーション
	User  User  `gorm:"foreignKey:UserID" json:"-"`
	Actor *User `gorm:"foreignKey:ActorUserID" json:"-"`
}

// IsUnread 未読かどうか
func (n *Notification) IsUnread() bool {
	return n.ReadAt == nil
}

// UserNotificationState ユーザーごとのお知らせ閲覧状態（更新情報をいつまで読んだか）
type UserNotificationState struct {
	UserID     uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	InfoReadAt *time.Time `json:"info_read_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
