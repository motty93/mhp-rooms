package models

import (
	"github.com/google/uuid"
)

// ReportAttachment 通報の添付ファイルモデル
type ReportAttachment struct {
	BaseModel
	ReportID     uuid.UUID `gorm:"type:uuid;not null;index" json:"report_id"`
	FilePath     string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileType     string    `gorm:"type:varchar(50);not null" json:"file_type"`
	FileSize     int64     `gorm:"not null" json:"file_size"`
	OriginalName string    `gorm:"type:varchar(255);not null" json:"original_name"`

	// リレーション
	Report UserReport `gorm:"foreignKey:ReportID" json:"-"`
}

// TableName テーブル名を指定
func (ReportAttachment) TableName() string {
	return "report_attachments"
}
