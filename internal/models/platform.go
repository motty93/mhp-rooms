package models

import (
	"time"

	"github.com/google/uuid"
)

type Platform struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	DisplayOrder int       `gorm:"not null" json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relations
	GameVersions []GameVersion `gorm:"foreignKey:PlatformID" json:"game_versions,omitempty"`
}
