package models

import (
	"time"

	"github.com/google/uuid"
)

// UserFollow ユーザー間のフォロー関係を管理
type UserFollow struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FollowerUserID  uuid.UUID  `gorm:"type:uuid;not null" json:"follower_user_id"`
	FollowingUserID uuid.UUID  `gorm:"type:uuid;not null" json:"following_user_id"`
	Status          string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, accepted, rejected
	CreatedAt       time.Time  `json:"created_at"`
	AcceptedAt      *time.Time `json:"accepted_at"`

	// リレーション
	Follower  User `gorm:"foreignKey:FollowerUserID" json:"follower,omitempty"`
	Following User `gorm:"foreignKey:FollowingUserID" json:"following,omitempty"`
}

// TableName テーブル名を指定
func (UserFollow) TableName() string {
	return "user_follows"
}

// IsMutual 相互フォロー（フレンド）かどうかを判定
func (uf *UserFollow) IsMutual(reverseFollow *UserFollow) bool {
	if uf == nil || reverseFollow == nil {
		return false
	}
	return uf.Status == "accepted" && reverseFollow.Status == "accepted"
}

// Accept フォローリクエストを承認
func (uf *UserFollow) Accept() {
	uf.Status = "accepted"
	now := time.Now()
	uf.AcceptedAt = &now
}

// Reject フォローリクエストを拒否
func (uf *UserFollow) Reject() {
	uf.Status = "rejected"
}

// UserFollowStatus フォロー状態の定数
const (
	FollowStatusPending  = "pending"
	FollowStatusAccepted = "accepted"
	FollowStatusRejected = "rejected"
)
