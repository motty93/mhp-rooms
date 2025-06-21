package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User はユーザー情報を管理するモデル
type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupabaseUserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"supabase_user_id"`
	Email          string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username       *string   `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	DisplayName    string    `gorm:"type:varchar(100);not null" json:"display_name"`
	AvatarURL      *string   `gorm:"type:text" json:"avatar_url"`
	Bio            *string   `gorm:"type:text" json:"bio"`
	PSNOnlineID    *string   `gorm:"type:varchar(16)" json:"psn_online_id"`
	TwitterID      *string   `gorm:"type:varchar(15)" json:"twitter_id"`
	IsActive       bool      `gorm:"not null;default:true" json:"is_active"`
	Role           string    `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// リレーション
	HostedRooms  []Room        `gorm:"foreignKey:HostUserID" json:"hosted_rooms,omitempty"`
	RoomMembers  []RoomMember  `gorm:"foreignKey:UserID" json:"room_members,omitempty"`
	Messages     []RoomMessage `gorm:"foreignKey:UserID" json:"messages,omitempty"`
	RoomLogs     []RoomLog     `gorm:"foreignKey:UserID" json:"room_logs,omitempty"`
	BlockedUsers []UserBlock   `gorm:"foreignKey:BlockerUserID" json:"blocked_users,omitempty"`
	BlockedBy    []UserBlock   `gorm:"foreignKey:BlockedUserID" json:"blocked_by,omitempty"`
}

// BeforeCreate はレコード作成前にUUIDを生成
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}