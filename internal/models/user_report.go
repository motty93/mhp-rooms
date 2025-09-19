package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ReportStatus 通報のステータス
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"   // 未対応
	ReportStatusReviewing ReportStatus = "reviewing" // 確認中
	ReportStatusResolved  ReportStatus = "resolved"  // 解決済み
	ReportStatusRejected  ReportStatus = "rejected"  // 却下
)

// ReportReason 通報理由
type ReportReason string

const (
	ReasonSpam            ReportReason = "spam"               // スパム・迷惑行為
	ReasonHarassment      ReportReason = "harassment"         // 嫌がらせ・誹謗中傷
	ReasonImpersonation   ReportReason = "impersonation"      // なりすまし
	ReasonInappropriate   ReportReason = "inappropriate"      // 不適切なコンテンツ
	ReasonScam            ReportReason = "scam"               // 詐欺・フィッシング
	ReasonPrivacyViolation ReportReason = "privacy_violation" // プライバシー侵害
	ReasonCheating        ReportReason = "cheating"           // チート行為
	ReasonOffensive       ReportReason = "offensive"          // 公序良俗違反
	ReasonOther           ReportReason = "other"              // その他
)

// ReportReasons 複数の通報理由を保持する型
type ReportReasons []ReportReason

// Value データベースに保存する際の値を返す
func (r ReportReasons) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

// Scan データベースから読み込む際の処理
func (r *ReportReasons) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		bytes = []byte(value.(string))
	}

	return json.Unmarshal(bytes, r)
}

// UserReport ユーザー通報モデル
type UserReport struct {
	BaseModel
	ReporterUserID uuid.UUID     `gorm:"type:uuid;not null;index" json:"reporter_user_id"`
	ReportedUserID uuid.UUID     `gorm:"type:uuid;not null;index" json:"reported_user_id"`
	Reasons        ReportReasons `gorm:"type:text;not null" json:"reasons"`
	Description    string        `gorm:"type:text;not null" json:"description"`
	Status         ReportStatus  `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	AdminNote      *string       `gorm:"type:text" json:"admin_note"`
	ResolvedAt     *time.Time    `json:"resolved_at"`

	// リレーション
	Reporter    User               `gorm:"foreignKey:ReporterUserID" json:"reporter"`
	Reported    User               `gorm:"foreignKey:ReportedUserID" json:"reported"`
	Attachments []ReportAttachment `gorm:"foreignKey:ReportID" json:"attachments"`
}

// TableName テーブル名を指定
func (UserReport) TableName() string {
	return "user_reports"
}

// GetReasonLabels 通報理由の日本語ラベルを返す
func GetReasonLabels() map[ReportReason]string {
	return map[ReportReason]string{
		ReasonSpam:             "スパム・迷惑行為",
		ReasonHarassment:       "嫌がらせ・誹謗中傷",
		ReasonImpersonation:    "なりすまし",
		ReasonInappropriate:    "不適切なコンテンツ",
		ReasonScam:             "詐欺・フィッシング",
		ReasonPrivacyViolation: "プライバシー侵害",
		ReasonCheating:         "チート行為",
		ReasonOffensive:        "公序良俗違反",
		ReasonOther:            "その他",
	}
}

// GetReasonDescription 通報理由の説明を返す
func GetReasonDescription() map[ReportReason]string {
	return map[ReportReason]string{
		ReasonSpam:             "繰り返しの宣伝、無関係な投稿",
		ReasonHarassment:       "暴言、脅迫、いじめ",
		ReasonImpersonation:    "他人を装う行為",
		ReasonInappropriate:    "規約違反のコンテンツ投稿",
		ReasonScam:             "詐欺的な行為、個人情報の不正取得",
		ReasonPrivacyViolation: "無断で個人情報を公開",
		ReasonCheating:         "ゲーム内での不正行為、改造、ツール使用",
		ReasonOffensive:        "わいせつ、暴力的、差別的な内容",
		ReasonOther:            "上記以外の問題",
	}
}