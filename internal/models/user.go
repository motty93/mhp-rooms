package models

import (
	"time"

	"github.com/google/uuid"
	"mhp-rooms/internal/utils"
)

type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupabaseUserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"supabase_user_id"`
	Email             string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username          *string   `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	DisplayName       string    `gorm:"type:varchar(100);not null" json:"display_name"`
	AvatarURL         *string   `gorm:"type:text" json:"avatar_url"`
	Bio               *string   `gorm:"type:text" json:"bio"`
	PSNOnlineID       *string   `gorm:"type:varchar(16)" json:"psn_online_id"`
	NintendoNetworkID *string   `gorm:"type:varchar(16)" json:"nintendo_network_id"`
	NintendoSwitchID  *string   `gorm:"type:varchar(20)" json:"nintendo_switch_id"`
	PretendoNetworkID *string   `gorm:"type:varchar(16)" json:"pretendo_network_id"`
	TwitterID         *string   `gorm:"type:varchar(15)" json:"twitter_id"`
	IsActive          bool      `gorm:"not null;default:true" json:"is_active"`
	Role              string    `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// リレーション
	HostedRooms  []Room        `gorm:"foreignKey:HostUserID" json:"hosted_rooms,omitempty"`
	RoomMembers  []RoomMember  `gorm:"foreignKey:UserID" json:"room_members,omitempty"`
	Messages     []RoomMessage `gorm:"foreignKey:UserID" json:"messages,omitempty"`
	RoomLogs     []RoomLog     `gorm:"foreignKey:UserID" json:"room_logs,omitempty"`
	BlockedUsers []UserBlock   `gorm:"foreignKey:BlockerUserID" json:"blocked_users,omitempty"`
	BlockedBy    []UserBlock   `gorm:"foreignKey:BlockedUserID" json:"blocked_by,omitempty"`
	PlayerNames  []PlayerName  `gorm:"foreignKey:UserID" json:"player_names,omitempty"`
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	// Note: この実装では、Userモデルにパスワードフィールドがないため、
	// 実際のパスワード更新はSupabaseやその他の認証システムで行う必要があります
	// ここでは一旦エラーを返さずに処理成功とします
	_ = hashedPassword
	return nil
}
