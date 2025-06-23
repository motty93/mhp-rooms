package models

import (
	"time"

	"github.com/google/uuid"
)

type GameVersion struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code         string    `gorm:"type:varchar(10);uniqueIndex;not null" json:"code"`
	Name         string    `gorm:"type:varchar(50);not null" json:"name"`
	DisplayOrder int       `gorm:"not null" json:"display_order"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// リレーション
	Rooms []Room `gorm:"foreignKey:GameVersionID" json:"rooms,omitempty"`
}
