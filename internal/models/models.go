package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User はユーザー情報を管理するモデル
type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupabaseUserID  uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"supabase_user_id"`
	Email           string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username        *string   `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	DisplayName     string    `gorm:"type:varchar(100);not null" json:"display_name"`
	AvatarURL       *string   `gorm:"type:text" json:"avatar_url"`
	Bio             *string   `gorm:"type:text" json:"bio"`
	IsActive        bool      `gorm:"not null;default:true" json:"is_active"`
	Role            string    `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// リレーション
	HostedRooms  []Room       `gorm:"foreignKey:HostUserID" json:"hosted_rooms,omitempty"`
	RoomMembers  []RoomMember `gorm:"foreignKey:UserID" json:"room_members,omitempty"`
	Messages     []RoomMessage `gorm:"foreignKey:UserID" json:"messages,omitempty"`
	RoomLogs     []RoomLog    `gorm:"foreignKey:UserID" json:"room_logs,omitempty"`
	BlockedUsers []UserBlock  `gorm:"foreignKey:BlockerUserID" json:"blocked_users,omitempty"`
	BlockedBy    []UserBlock  `gorm:"foreignKey:BlockedUserID" json:"blocked_by,omitempty"`
}

// GameVersion はゲームバージョンマスターを管理するモデル
type GameVersion struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code         string    `gorm:"type:varchar(10);uniqueIndex;not null" json:"code"`
	Name         string    `gorm:"type:varchar(50);not null" json:"name"`
	DisplayOrder int       `gorm:"not null" json:"display_order"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`

	// リレーション
	Rooms []Room `gorm:"foreignKey:GameVersionID" json:"rooms,omitempty"`
}

// Room はルーム情報を管理するモデル
type Room struct {
	ID               uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoomCode         string       `gorm:"type:varchar(20);uniqueIndex;not null" json:"room_code"`
	Name             string       `gorm:"type:varchar(100);not null" json:"name"`
	Description      *string      `gorm:"type:text" json:"description"`
	GameVersionID    uuid.UUID    `gorm:"type:uuid;not null" json:"game_version_id"`
	HostUserID       uuid.UUID    `gorm:"type:uuid;not null" json:"host_user_id"`
	MaxPlayers       int          `gorm:"not null;default:4" json:"max_players"`
	CurrentPlayers   int          `gorm:"not null;default:0" json:"current_players"`
	PasswordHash     *string      `gorm:"type:varchar(255)" json:"password_hash,omitempty"`
	Status           string       `gorm:"type:varchar(20);not null;default:'waiting'" json:"status"`
	QuestType        *string      `gorm:"type:varchar(50)" json:"quest_type"`
	TargetMonster    *string      `gorm:"type:varchar(100)" json:"target_monster"`
	RankRequirement  *string      `gorm:"type:varchar(20)" json:"rank_requirement"`
	IsActive         bool         `gorm:"not null;default:true" json:"is_active"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	ClosedAt         *time.Time   `json:"closed_at"`

	// リレーション
	GameVersion GameVersion   `gorm:"foreignKey:GameVersionID" json:"game_version"`
	Host        User          `gorm:"foreignKey:HostUserID" json:"host"`
	Members     []RoomMember  `gorm:"foreignKey:RoomID" json:"members,omitempty"`
	Messages    []RoomMessage `gorm:"foreignKey:RoomID" json:"messages,omitempty"`
	Logs        []RoomLog     `gorm:"foreignKey:RoomID" json:"logs,omitempty"`
}

// RoomMember はルームメンバーを管理するモデル
type RoomMember struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoomID       uuid.UUID  `gorm:"type:uuid;not null" json:"room_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	PlayerNumber int        `gorm:"not null" json:"player_number"`
	IsHost       bool       `gorm:"not null;default:false" json:"is_host"`
	Status       string     `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	JoinedAt     time.Time  `json:"joined_at"`
	LeftAt       *time.Time `json:"left_at"`

	// リレーション
	Room Room `gorm:"foreignKey:RoomID" json:"room"`
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// RoomMessage はルームメッセージを管理するモデル
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

// UserBlock はユーザーブロックを管理するモデル
type UserBlock struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BlockerUserID uuid.UUID  `gorm:"type:uuid;not null" json:"blocker_user_id"`
	BlockedUserID uuid.UUID  `gorm:"type:uuid;not null" json:"blocked_user_id"`
	Reason        *string    `gorm:"type:text" json:"reason"`
	CreatedAt     time.Time  `json:"created_at"`

	// リレーション
	Blocker User `gorm:"foreignKey:BlockerUserID" json:"blocker"`
	Blocked User `gorm:"foreignKey:BlockedUserID" json:"blocked"`
}

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
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (gv *GameVersion) BeforeCreate(tx *gorm.DB) error {
	if gv.ID == uuid.Nil {
		gv.ID = uuid.New()
	}
	return nil
}

func (r *Room) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (rm *RoomMember) BeforeCreate(tx *gorm.DB) error {
	if rm.ID == uuid.Nil {
		rm.ID = uuid.New()
	}
	return nil
}

func (rmsg *RoomMessage) BeforeCreate(tx *gorm.DB) error {
	if rmsg.ID == uuid.Nil {
		rmsg.ID = uuid.New()
	}
	return nil
}

func (ub *UserBlock) BeforeCreate(tx *gorm.DB) error {
	if ub.ID == uuid.Nil {
		ub.ID = uuid.New()
	}
	return nil
}

func (rl *RoomLog) BeforeCreate(tx *gorm.DB) error {
	if rl.ID == uuid.Nil {
		rl.ID = uuid.New()
	}
	return nil
}