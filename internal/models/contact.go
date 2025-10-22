package models

import (
	"github.com/google/uuid"
)

// Contact お問合せモデル
type Contact struct {
	BaseModel
	InquiryType     string     `gorm:"type:varchar(50);not null" json:"inquiry_type"`     // お問合せ種類
	Name            string     `gorm:"type:varchar(100);not null" json:"name"`            // お名前
	Email           string     `gorm:"type:varchar(255);not null" json:"email"`           // メールアドレス
	Subject         string     `gorm:"type:varchar(200);not null" json:"subject"`         // 件名
	Message         string     `gorm:"type:text;not null" json:"message"`                 // お問い合わせ内容
	IPAddress       string     `gorm:"type:varchar(45)" json:"ip_address"`                // IPアドレス（IPv6対応で45文字）
	UserAgent       string     `gorm:"type:text" json:"user_agent"`                       // User Agent
	IsAuthenticated bool       `gorm:"default:false" json:"is_authenticated"`             // 認証済みユーザーかどうか
	SupabaseUserID  *uuid.UUID `gorm:"type:uuid;index" json:"supabase_user_id,omitempty"` // Supabase User ID（認証済みの場合）
}

// TableName Contact テーブル名を指定
func (Contact) TableName() string {
	return "contacts"
}
