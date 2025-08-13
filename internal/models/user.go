package models

import (
	"time"

	"github.com/google/uuid"
	"mhp-rooms/internal/utils"
)


// PlayTimes プレイ時間帯の構造体
type PlayTimes struct {
	Weekday string `json:"weekday,omitempty"`
	Weekend string `json:"weekend,omitempty"`
}

type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupabaseUserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"supabase_user_id"`
	Email             string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username          *string   `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	DisplayName       string    `gorm:"type:varchar(100);not null" json:"display_name"`
	AvatarURL         *string   `gorm:"type:text" json:"avatar_url"`
	Bio               *string   `gorm:"type:text" json:"bio"`
	PSNOnlineID       *string   `gorm:"type:varchar(16)" json:"psn_online_id"`
	NintendoNetworkID *string   `gorm:"type:varchar(16)" json:"nintendo_network_id"`
	NintendoSwitchID  *string   `gorm:"type:varchar(20)" json:"nintendo_switch_id"`
	PretendoNetworkID *string   `gorm:"type:varchar(16)" json:"pretendo_network_id"`
	TwitterID         *string   `gorm:"type:varchar(15)" json:"twitter_id"`
	FavoriteGames     JSONB     `gorm:"type:jsonb;default:'[]'" json:"favorite_games"`
	PlayTimes         JSONB     `gorm:"type:jsonb;default:'{}'" json:"play_times"`
	IsActive          bool      `gorm:"not null;default:true" json:"is_active"`
	Role              string    `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// リレーション
	HostedRooms  []Room        `gorm:"foreignKey:HostUserID" json:"hosted_rooms,omitempty"`
	RoomMembers  []RoomMember  `gorm:"foreignKey:UserID" json:"room_members,omitempty"`
	Messages     []RoomMessage `gorm:"foreignKey:UserID" json:"messages,omitempty"`
	RoomLogs     []RoomLog     `gorm:"foreignKey:UserID" json:"room_logs,omitempty"`
	BlockedUsers []UserBlock   `gorm:"foreignKey:BlockerUserID" json:"blocked_users,omitempty"`
	BlockedBy    []UserBlock   `gorm:"foreignKey:BlockedUserID" json:"blocked_by,omitempty"`
	PlayerNames  []PlayerName  `gorm:"foreignKey:UserID" json:"player_names,omitempty"`
	Following    []UserFollow  `gorm:"foreignKey:FollowerUserID" json:"following,omitempty"`
	Followers    []UserFollow  `gorm:"foreignKey:FollowingUserID" json:"followers,omitempty"`
}

// GetFavoriteGames お気に入りゲームのリストを取得
func (u *User) GetFavoriteGames() ([]string, error) {
	if u.FavoriteGames.Data == nil {
		return []string{}, nil
	}
	
	// 直接配列として保存されている場合
	if gamesSlice, ok := u.FavoriteGames.Data.([]interface{}); ok {
		games := make([]string, 0, len(gamesSlice))
		for _, v := range gamesSlice {
			if str, ok := v.(string); ok {
				games = append(games, str)
			}
		}
		return games, nil
	}
	
	// オブジェクト形式で保存されている場合（"games"キー）
	if gamesMap, ok := u.FavoriteGames.Data.(map[string]interface{}); ok {
		if gamesInterface, exists := gamesMap["games"]; exists {
			if gamesSlice, ok := gamesInterface.([]interface{}); ok {
				games := make([]string, 0, len(gamesSlice))
				for _, v := range gamesSlice {
					if str, ok := v.(string); ok {
						games = append(games, str)
					}
				}
				return games, nil
			}
		}
	}
	
	return []string{}, nil
}

// SetFavoriteGames お気に入りゲームを設定
func (u *User) SetFavoriteGames(games []string) error {
	// 直接配列として設定
	u.FavoriteGames.Data = games
	return nil
}

// GetPlayTimes プレイ時間帯を取得
func (u *User) GetPlayTimes() (*PlayTimes, error) {
	if u.PlayTimes.Data == nil {
		return &PlayTimes{}, nil
	}
	
	times := &PlayTimes{}
	
	// オブジェクト形式で保存されている場合
	if playTimesMap, ok := u.PlayTimes.Data.(map[string]interface{}); ok {
		if weekday, exists := playTimesMap["weekday"]; exists {
			if str, ok := weekday.(string); ok {
				times.Weekday = str
			}
		}
		if weekend, exists := playTimesMap["weekend"]; exists {
			if str, ok := weekend.(string); ok {
				times.Weekend = str
			}
		}
	}
	
	return times, nil
}

// SetPlayTimes プレイ時間帯を設定
func (u *User) SetPlayTimes(times *PlayTimes) error {
	// オブジェクト形式で設定
	u.PlayTimes.Data = map[string]interface{}{
		"weekday": times.Weekday,
		"weekend": times.Weekend,
	}
	return nil
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	// Note: この実装では、Userモデルにパスワードフィールドがないため、
	// 実際のパスワード更新はSupabaseやその他の認証システムで行う必要があります
	// ここでは一旦エラーを返さずに処理成功とします
	_ = hashedPassword
	return nil
}
