package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserActivity ユーザーの行動履歴を記録するモデル
type UserActivity struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ActivityType      string     `gorm:"size:50;not null;index" json:"activity_type"`
	Title             string     `gorm:"size:255;not null" json:"title"`
	Description       *string    `gorm:"type:text" json:"description"`
	RelatedEntityType *string    `gorm:"size:50" json:"related_entity_type"`
	RelatedEntityID   *uuid.UUID `gorm:"type:uuid" json:"related_entity_id"`
	Metadata          JSONB      `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	Icon              string     `gorm:"size:100" json:"icon"`
	IconColor         string     `gorm:"size:50" json:"icon_color"`
	CreatedAt         time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// リレーション
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// アクティビティタイプ定数
const (
	// 部屋関連
	ActivityRoomCreate = "room_create" // 部屋作成
	ActivityRoomJoin   = "room_join"   // 部屋参加
	ActivityRoomLeave  = "room_leave"  // 部屋退出
	ActivityRoomClose  = "room_close"  // 部屋終了
	ActivityRoomUpdate = "room_update" // 部屋設定変更

	// フォロー関連
	ActivityFollowAdd    = "follow_add"    // フォロー開始
	ActivityFollowAccept = "follow_accept" // フォロー承認
	ActivityFollowRemove = "follow_remove" // フォロー解除

	// メッセージ関連
	ActivityMessageSend = "message_send" // メッセージ送信

	// プロフィール関連
	ActivityProfileUpdate = "profile_update" // プロフィール更新

	// システム関連
	ActivityUserJoin = "user_join" // ユーザー登録
)

// エンティティタイプ定数
const (
	EntityTypeRoom    = "room"
	EntityTypeUser    = "user"
	EntityTypeMessage = "message"
)

// アクティビティメタデータの構造体定義

// RoomActivityMetadata 部屋関連アクティビティのメタデータ
type RoomActivityMetadata struct {
	GameVersion    string `json:"game_version,omitempty"`
	MaxPlayers     int    `json:"max_players,omitempty"`
	TargetMonster  string `json:"target_monster,omitempty"`
	HostUserID     string `json:"host_user_id,omitempty"`
	RoomPassword   bool   `json:"room_password,omitempty"` // パスワード有無（実際のパスワードは保存しない）
}

// FollowActivityMetadata フォロー関連アクティビティのメタデータ
type FollowActivityMetadata struct {
	FollowingUserID string `json:"following_user_id,omitempty"`
	FollowerUserID  string `json:"follower_user_id,omitempty"`
	IsMutualFollow  bool   `json:"is_mutual_follow,omitempty"`
}

// MessageActivityMetadata メッセージ関連アクティビティのメタデータ
type MessageActivityMetadata struct {
	RoomID      string `json:"room_id,omitempty"`
	MessageType string `json:"message_type,omitempty"`
}

// UserJoinActivityMetadata ユーザー登録関連アクティビティのメタデータ
type UserJoinActivityMetadata struct {
	RegistrationMethod string `json:"registration_method,omitempty"` // email, google, githubなど
}

// GetMetadata メタデータをJSONから指定した構造体に変換
func (ua *UserActivity) GetMetadata(target interface{}) error {
	if ua.Metadata.Data == nil {
		return nil
	}

	jsonData, err := json.Marshal(ua.Metadata.Data)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, target)
}

// SetMetadata 指定した構造体をJSONBメタデータとして設定
func (ua *UserActivity) SetMetadata(data interface{}) error {
	ua.Metadata.Data = data
	return nil
}

// TableName テーブル名を指定
func (UserActivity) TableName() string {
	return "user_activities"
}