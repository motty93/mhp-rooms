package models

type Platform struct {
	BaseModel
	Name         string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	DisplayOrder int    `gorm:"not null" json:"display_order"`

	// Relations
	GameVersions []GameVersion `gorm:"foreignKey:PlatformID" json:"game_versions,omitempty"`
}
