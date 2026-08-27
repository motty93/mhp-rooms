package repository

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
)

// userRepository はユーザー関連の操作を行うリポジトリの実装
type userRepository struct {
	db DBInterface
}

// PublicHunterListParams は公開ハンター一覧の検索条件です。
type PublicHunterListParams struct {
	Query, Sort   string
	Limit, Offset int
}

// PublicHunter は公開ハンター一覧カードに必要な最小限の表示データです。
type PublicHunter struct {
	ID                  uuid.UUID
	DisplayName         string
	Username, AvatarURL *string
	CreatedAt           time.Time
	RecentActivityTitle *string
	RecentActivityAt    string
	RoomCreateCount     int64
}

// AdminUserListParams は管理画面のユーザー一覧の検索・並び替え条件です。
type AdminUserListParams struct {
	Query, Sort   string
	Limit, Offset int
}

// AdminUser は管理画面だけで使用するユーザー一覧の表示データです。
// Email は管理者向けテンプレートにのみ渡します。
type AdminUser struct {
	ID             uuid.UUID
	DisplayName    string
	Username       *string
	Email          string
	Role           string
	CreatedAt      time.Time
	RoomCount      int64
	ReportCount    int64
	LastActivityAt *time.Time
}

// adminUserRecord はDBからの集計結果を受け取るための内部表現です。
// SQLite は MAX(created_at) を文字列として返すため、最終活動日時には専用の Scanner を使います。
type adminUserRecord struct {
	ID             uuid.UUID
	DisplayName    string
	Username       *string
	Email          string
	Role           string
	CreatedAt      time.Time
	RoomCount      int64
	ReportCount    int64
	LastActivityAt nullableAdminTime
}

type nullableAdminTime struct {
	Time  time.Time
	Valid bool
}

func (t nullableAdminTime) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}
	return t.Time, nil
}

func (t *nullableAdminTime) Scan(value interface{}) error {
	t.Time = time.Time{}
	t.Valid = false
	if value == nil {
		return nil
	}

	var raw string
	switch value := value.(type) {
	case time.Time:
		t.Time = value
		t.Valid = true
		return nil
	case string:
		raw = value
	case []byte:
		raw = string(value)
	default:
		return fmt.Errorf("scan admin activity time %T", value)
	}

	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	} {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			t.Time = parsed
			t.Valid = true
			return nil
		}
	}
	return fmt.Errorf("parse admin activity time %q", raw)
}

// NewUserRepository は新しいUserRepositoryインスタンスを作成
func NewUserRepository(db DBInterface) UserRepository {
	return &userRepository{db: db}
}

// CreateUser はユーザーを作成
func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.GetConn().Create(user).Error
}

// FindUserByID はIDでユーザーを検索
func (r *userRepository) FindUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().
		Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "favorite_games", "play_times", "is_active", "role", "created_at", "updated_at").
		Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, err
	}
	return &user, nil
}

// FindUsersByIDs は複数のIDでユーザーを一括取得
func (r *userRepository) FindUsersByIDs(ids []uuid.UUID) ([]models.User, error) {
	var users []models.User
	if len(ids) == 0 {
		return users, nil
	}

	err := r.db.GetConn().Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// FindUserBySupabaseUserID はSupabaseユーザーIDでユーザーを検索
func (r *userRepository) FindUserBySupabaseUserID(supabaseUserID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().
		Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "favorite_games", "play_times", "is_active", "role", "created_at", "updated_at").
		Where("supabase_user_id = ?", supabaseUserID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail はメールアドレスでユーザーを検索
func (r *userRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().
		Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "favorite_games", "play_times", "is_active", "role", "created_at", "updated_at").
		Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser はユーザー情報を更新
func (r *userRepository) UpdateUser(user *models.User) error {
	// Turso対応：明示的なトランザクション処理
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetActiveUsers はアクティブなユーザー一覧を取得
func (r *userRepository) GetActiveUsers(limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.GetConn().
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

// ListAdminUsers は無効化済みを含む管理用ユーザー一覧を集計付きで返します。
func (r *userRepository) ListAdminUsers(params AdminUserListParams) ([]AdminUser, error) {
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 20
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	if params.Sort != "rooms" && params.Sort != "reports" && params.Sort != "joined" {
		params.Sort = "joined"
	}

	roomCount := "(SELECT COUNT(*) FROM rooms admin_rooms WHERE admin_rooms.host_user_id = users.id)"
	reportCount := "(SELECT COUNT(*) FROM user_reports admin_reports WHERE admin_reports.reported_user_id = users.id)"
	lastActivityAt := "(SELECT MAX(admin_activities.created_at) FROM user_activities admin_activities WHERE admin_activities.user_id = users.id)"
	query := r.db.GetConn().Model(&models.User{}).Select(
		"users.id, users.display_name, users.username, users.email, users.role, users.created_at, " + roomCount + " AS room_count, " + reportCount + " AS report_count, " + lastActivityAt + " AS last_activity_at",
	)
	if normalized := strings.TrimSpace(params.Query); normalized != "" {
		like := "%" + strings.ToLower(normalized) + "%"
		query = query.Where("(LOWER(users.display_name) LIKE ? OR LOWER(COALESCE(users.username, '')) LIKE ? OR LOWER(users.email) LIKE ?)", like, like, like)
	}
	switch params.Sort {
	case "rooms":
		query = query.Order("room_count DESC, users.created_at DESC, users.id ASC")
	case "reports":
		query = query.Order("report_count DESC, users.created_at DESC, users.id ASC")
	default:
		query = query.Order("users.created_at DESC, users.id ASC")
	}

	var records []adminUserRecord
	if err := query.Limit(params.Limit).Offset(params.Offset).Scan(&records).Error; err != nil {
		return nil, err
	}
	users := make([]AdminUser, 0, len(records))
	for _, record := range records {
		var lastActivityAt *time.Time
		if record.LastActivityAt.Valid {
			value := record.LastActivityAt.Time
			lastActivityAt = &value
		}
		users = append(users, AdminUser{
			ID:             record.ID,
			DisplayName:    record.DisplayName,
			Username:       record.Username,
			Email:          record.Email,
			Role:           record.Role,
			CreatedAt:      record.CreatedAt,
			RoomCount:      record.RoomCount,
			ReportCount:    record.ReportCount,
			LastActivityAt: lastActivityAt,
		})
	}
	return users, nil
}

// CountAdminUsers は管理画面の検索条件に一致する全ユーザー数を返します。
func (r *userRepository) CountAdminUsers(search string) (int64, error) {
	query := r.db.GetConn().Model(&models.User{})
	if normalized := strings.TrimSpace(search); normalized != "" {
		like := "%" + strings.ToLower(normalized) + "%"
		query = query.Where("(LOWER(display_name) LIKE ? OR LOWER(COALESCE(username, '')) LIKE ? OR LOWER(email) LIKE ?)", like, like, like)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ListPublicHunters は公開ハンター一覧をN+1なしで取得します。
func (r *userRepository) ListPublicHunters(params PublicHunterListParams) ([]PublicHunter, error) {
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 20
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	if params.Sort != "rooms" && params.Sort != "joined" {
		params.Sort = "recent"
	}
	recentAt := "COALESCE((SELECT MAX(ua_recent.created_at) FROM user_activities ua_recent WHERE ua_recent.user_id = users.id AND ua_recent.activity_type IN ?), users.created_at)"
	roomCount := "(SELECT COUNT(*) FROM user_activities ua_rooms WHERE ua_rooms.user_id = users.id AND ua_rooms.activity_type = ?)"
	recentTitle := "(SELECT ua_title.title FROM user_activities ua_title WHERE ua_title.user_id = users.id AND ua_title.activity_type IN ? ORDER BY ua_title.created_at DESC LIMIT 1)"
	query := r.db.GetConn().Model(&models.User{}).
		Select(
			"users.id, users.display_name, users.username, users.avatar_url, users.created_at, "+recentTitle+" AS recent_activity_title, "+recentAt+" AS recent_activity_at, "+roomCount+" AS room_create_count",
			models.PublicFeedActivityTypes(),
			models.PublicFeedActivityTypes(),
			models.ActivityRoomCreate,
		).
		Where("users.is_active = ?", true)
	if normalized := strings.TrimSpace(params.Query); normalized != "" {
		like := "%" + strings.ToLower(normalized) + "%"
		query = query.Where("(LOWER(users.display_name) LIKE ? OR LOWER(COALESCE(users.username, '')) LIKE ?)", like, like)
	}
	switch params.Sort {
	case "rooms":
		query = query.Order("room_create_count DESC, users.created_at DESC, users.id ASC")
	case "joined":
		query = query.Order("users.created_at DESC, users.id ASC")
	default:
		query = query.Order("recent_activity_at DESC, users.created_at DESC, users.id ASC")
	}
	var hunters []PublicHunter
	err := query.Limit(params.Limit).Offset(params.Offset).Scan(&hunters).Error
	return hunters, err
}

// CountPublicHunters は検索条件に一致する有効な公開ハンター数を返します。
func (r *userRepository) CountPublicHunters(search string) (int64, error) {
	query := r.db.GetConn().Model(&models.User{}).Where("is_active = ?", true)
	if normalized := strings.TrimSpace(search); normalized != "" {
		like := "%" + strings.ToLower(normalized) + "%"
		query = query.Where("(LOWER(display_name) LIKE ? OR LOWER(COALESCE(username, '')) LIKE ?)", like, like)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}
