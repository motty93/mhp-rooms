package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/utils"
)

type Room struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoomCode        string     `gorm:"type:varchar(20);uniqueIndex;not null" json:"room_code"`
	Name            string     `gorm:"type:varchar(100);not null" json:"name"`
	Description     *string    `gorm:"type:text" json:"description"`
	GameVersionID   uuid.UUID  `gorm:"type:uuid;not null" json:"game_version_id"`
	HostUserID      uuid.UUID  `gorm:"type:uuid;not null" json:"host_user_id"`
	MaxPlayers      int        `gorm:"not null;default:4" json:"max_players"`
	CurrentPlayers  int        `gorm:"not null;default:0" json:"current_players"`
	PasswordHash    *string    `gorm:"type:varchar(255)" json:"password_hash,omitempty"`
	Status          string     `gorm:"type:varchar(20);not null;default:'waiting'" json:"status"`
	QuestType       *string    `gorm:"type:varchar(50)" json:"quest_type"`
	TargetMonster   *string    `gorm:"type:varchar(100)" json:"target_monster"`
	RankRequirement *string    `gorm:"type:varchar(20)" json:"rank_requirement"`
	IsActive        bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ClosedAt        *time.Time `json:"closed_at"`

	// リレーション
	GameVersion GameVersion   `gorm:"foreignKey:GameVersionID" json:"game_version"`
	Host        User          `gorm:"foreignKey:HostUserID" json:"host"`
	Members     []RoomMember  `gorm:"foreignKey:RoomID" json:"members,omitempty"`
	Messages    []RoomMessage `gorm:"foreignKey:RoomID" json:"messages,omitempty"`
	Logs        []RoomLog     `gorm:"foreignKey:RoomID" json:"logs,omitempty"`
}

func (r *Room) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (r *Room) SetPassword(password string) error {
	if password == "" {
		r.PasswordHash = nil
		return nil
	}
	
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	r.PasswordHash = &hash
	return nil
}

func (r *Room) CheckPassword(password string) bool {
	if r.PasswordHash == nil {
		return password == ""
	}
	return utils.CheckPassword(password, *r.PasswordHash)
}

func (r *Room) HasPassword() bool {
	return r.PasswordHash != nil
}

func (r *Room) IsFull() bool {
	return r.CurrentPlayers >= r.MaxPlayers
}

func (r *Room) CanJoin() bool {
	return r.IsActive && r.Status == "waiting" && !r.IsFull()
}