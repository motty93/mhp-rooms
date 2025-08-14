package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

type ProfileHandler struct {
	BaseHandler
	logger *log.Logger
}

func NewProfileHandler(repo *repository.Repository) *ProfileHandler {
	return &ProfileHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		logger: log.New(log.Writer(), "[ProfileHandler] ", log.LstdFlags),
	}
}

// ProfileData プロフィールページ用のデータ構造
type ProfileData struct {
	User          *models.User      `json:"user"`
	IsOwnProfile  bool              `json:"isOwnProfile"`
	Activities    []Activity        `json:"activities"`
	Rooms         []RoomSummary     `json:"rooms"`
	Followers     []Follower        `json:"followers"`
	FollowerCount int64             `json:"followerCount"`
	FavoriteGames []string          `json:"favoriteGames"`
	PlayTimes     *models.PlayTimes `json:"playTimes"`
}

// Activity アクティビティ情報
type Activity struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	TimeAgo     string `json:"timeAgo"`
	Icon        string `json:"icon"`
	IconColor   string `json:"iconColor"`
}

// RoomSummary 部屋の概要情報
type RoomSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	GameVersion string    `json:"gameVersion"`
	PlayerCount string    `json:"playerCount"`
	Status      string    `json:"status"`
	StatusColor string    `json:"statusColor"`
	CreatedAt   string    `json:"createdAt"`
	IsClickable bool      `json:"isClickable"`
}

// Follower フォロワー情報
type Follower struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	AvatarURL      string    `json:"avatarUrl"`
	IsOnline       bool      `json:"isOnline"`
	FollowingSince string    `json:"followingSince"`
}

// Profile 自分のプロフィールページを表示
func (ph *ProfileHandler) Profile(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		// 認証が必要 - 開発環境も本番環境も同じ処理
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	favoriteGames, _ := user.GetFavoriteGames()
	playTimes, _ := user.GetPlayTimes()

	// フォロワー数を取得（開発環境では25人固定）
	var followerCount int64 = 25
	if ph.repo != nil && ph.repo.UserFollow != nil {
		followers, err := ph.repo.UserFollow.GetFollowers(user.ID)
		if err == nil {
			followerCount = int64(len(followers))
		}
	}

	// 実際に作成した部屋を取得
	rooms, err := ph.repo.Room.GetRoomsByHostUser(user.ID, 10, 0) // 最大10件取得
	var roomSummaries []RoomSummary
	if err == nil {
		for _, room := range rooms {
			roomSummaries = append(roomSummaries, roomToSummary(room))
		}
	}

	profileData := ProfileData{
		User:          user,
		IsOwnProfile:  true,
		Activities:    ph.getMockActivities(),
		Rooms:         roomSummaries,
		Followers:     ph.getMockFollowers(),
		FollowerCount: followerCount,
		FavoriteGames: favoriteGames,
		PlayTimes:     playTimes,
	}

	data := TemplateData{
		Title:    "プロフィール",
		PageData: profileData,
	}

	renderTemplate(w, "profile.tmpl", data)
}

// UserProfile 他のユーザーのプロフィールページを表示
func (ph *ProfileHandler) UserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "uuid")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
		return
	}

	// データベースからユーザー情報を取得
	user, err := ph.repo.User.FindUserByID(userID)
	if err != nil {
		http.Error(w, "ユーザーが見つかりません", http.StatusNotFound)
		return
	}

	// 現在のユーザーと比較して自分のプロフィールかどうか判定
	currentUser := getUserFromContext(r.Context())
	isOwnProfile := false
	if currentUser != nil && currentUser.ID == user.ID {
		isOwnProfile = true
	}

	// お気に入りゲームとプレイ時間帯を取得
	favoriteGames, _ := user.GetFavoriteGames()
	playTimes, _ := user.GetPlayTimes()

	// フォロワー数を取得
	var followerCount int64 = 25
	if ph.repo != nil && ph.repo.UserFollow != nil {
		followers, err := ph.repo.UserFollow.GetFollowers(user.ID)
		if err == nil {
			followerCount = int64(len(followers))
		}
	}

	// 実際に作成した部屋を取得
	rooms, err := ph.repo.Room.GetRoomsByHostUser(user.ID, 10, 0) // 最大10件取得
	var roomSummaries []RoomSummary
	if err == nil {
		for _, room := range rooms {
			roomSummaries = append(roomSummaries, roomToSummary(room))
		}
	}

	profileData := ProfileData{
		User:          user,
		IsOwnProfile:  isOwnProfile,
		Activities:    ph.getMockActivities(),
		Rooms:         roomSummaries,
		Followers:     ph.getMockFollowers(),
		FollowerCount: followerCount,
		FavoriteGames: favoriteGames,
		PlayTimes:     playTimes,
	}

	data := TemplateData{
		Title:    user.DisplayName + "のプロフィール",
		PageData: profileData,
	}

	renderTemplate(w, "profile.tmpl", data)
}

// GetUserProfile API経由でユーザープロフィール情報を取得
func (ph *ProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "uuid")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "無効なユーザーIDです")
		return
	}

	user, err := ph.repo.User.FindUserByID(userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "ユーザーが見つかりません")
		return
	}

	// 現在のユーザーと比較
	currentUser := getUserFromContext(r.Context())
	isOwnProfile := false
	if currentUser != nil && currentUser.ID == user.ID {
		isOwnProfile = true
	}

	// お気に入りゲームとプレイ時間帯を取得
	favoriteGames, _ := user.GetFavoriteGames()
	playTimes, _ := user.GetPlayTimes()

	// フォロワー数を取得
	var followerCount int64 = 0
	if ph.repo != nil && ph.repo.UserFollow != nil {
		followers, err := ph.repo.UserFollow.GetFollowers(user.ID)
		if err == nil {
			followerCount = int64(len(followers))
		}
	}

	// 実際に作成した部屋を取得
	rooms, err := ph.repo.Room.GetRoomsByHostUser(user.ID, 10, 0) // 最大10件取得
	var roomSummaries []RoomSummary
	if err == nil {
		for _, room := range rooms {
			roomSummaries = append(roomSummaries, roomToSummary(room))
		}
	}

	profileData := ProfileData{
		User:          user,
		IsOwnProfile:  isOwnProfile,
		Activities:    ph.getMockActivities(),
		Rooms:         roomSummaries,
		Followers:     ph.getMockFollowers(),
		FollowerCount: followerCount,
		FavoriteGames: favoriteGames,
		PlayTimes:     playTimes,
	}

	respondWithJSON(w, http.StatusOK, profileData)
}

// EditForm プロフィール編集フォームを返す（htmx用）
func (ph *ProfileHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		// 認証が必要 - 開発環境も本番環境も同じ処理
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	bio := ""
	if user.Bio != nil {
		bio = *user.Bio
	}

	data := struct {
		Username string
		Bio      string
	}{
		Username: username,
		Bio:      bio,
	}

	if err := renderPartialTemplate(w, "profile_edit_form.tmpl", data); err != nil {
		ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
		http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
		return
	}
}

// Activity アクティビティタブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Activity(w http.ResponseWriter, r *http.Request) {
	// URLパラメータからユーザーIDを取得（他ユーザーのプロフィール表示用）
	var targetUserID uuid.UUID
	var err error

	userIDParam := chi.URLParam(r, "userID")
	if userIDParam != "" {
		targetUserID, err = uuid.Parse(userIDParam)
		if err != nil {
			ph.logger.Printf("無効なユーザーID: %s, エラー: %v", userIDParam, err)
			http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
			return
		}
	} else {
		// URLパラメータがない場合は自分のプロフィール
		dbUser, exists := middleware.GetDBUserFromContext(r.Context())
		if !exists || dbUser == nil {
			ph.logger.Printf("認証情報が取得できません")
			http.Error(w, "認証されていません", http.StatusUnauthorized)
			return
		}
		targetUserID = dbUser.ID
	}

	// データベースからアクティビティを取得
	userActivities, err := ph.repo.UserActivity.GetUserActivities(targetUserID, 20, 0)
	if err != nil {
		ph.logger.Printf("アクティビティ取得エラー: %v", err)
		// エラー時はフォールバック（空の配列を返す）
		userActivities = []models.UserActivity{}
	}

	// models.UserActivityをActivity構造体に変換
	displayActivities := make([]Activity, len(userActivities))
	for i, activity := range userActivities {
		displayActivities[i] = Activity{
			Type:        activity.ActivityType,
			Title:       activity.Title,
			Description: getStringValue(activity.Description),
			TimeAgo:     formatRelativeTime(activity.CreatedAt),
			Icon:        activity.Icon,
			IconColor:   activity.IconColor,
		}
	}

	data := struct {
		Activities []Activity
	}{
		Activities: displayActivities,
	}

	if err := renderPartialTemplate(w, "profile_activity.tmpl", data); err != nil {
		ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
		http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
		return
	}
}

// Rooms 作成した部屋タブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Rooms(w http.ResponseWriter, r *http.Request) {
	var targetUserID uuid.UUID
	var err error

	userIDStr := chi.URLParam(r, "uuid")
	if userIDStr != "" {
		// 他のユーザーのプロフィール
		targetUserID, err = uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
			return
		}
	} else {
		// 自分のプロフィール
		user := getUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}

		targetUserID = user.ID
	}

	// デバッグログ: どのユーザーIDで検索しているかを確認
	ph.logger.Printf("作成した部屋を検索中 - ユーザーID: %s", targetUserID.String())

	// ユーザーが作成した部屋を取得
	rooms, err := ph.repo.Room.GetRoomsByHostUser(targetUserID, 50, 0) // 最大50件取得
	if err != nil {
		ph.logger.Printf("部屋取得エラー: %v", err)
		http.Error(w, "部屋データの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	ph.logger.Printf("取得した部屋数: %d", len(rooms))

	// models.RoomをRoomSummaryに変換
	var roomSummaries []RoomSummary
	for _, room := range rooms {
		roomSummaries = append(roomSummaries, roomToSummary(room))
	}

	// テンプレートデータを準備
	data := struct {
		Rooms []RoomSummary
	}{
		Rooms: roomSummaries,
	}

	// 部分テンプレートを使用してレンダリング
	if err := renderPartialTemplate(w, "profile_rooms.tmpl", data); err != nil {
		ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
		http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
		return
	}
}

// Following フォロー中タブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Following(w http.ResponseWriter, r *http.Request) {
	if err := renderPartialTemplate(w, "profile_following.tmpl", nil); err != nil {
		ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
		http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
		return
	}
}

// Followers フォロワータブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Followers(w http.ResponseWriter, r *http.Request) {
	if err := renderPartialTemplate(w, "profile_followers.tmpl", nil); err != nil {
		ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
		http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
		return
	}
}

// Helper functions
func (ph *ProfileHandler) getMockActivities() []Activity {
	return []Activity{
		{
			Type:        "room_create",
			Title:       "【部屋作成】古龍種連戦",
			Description: "ターゲット: クシャルダオラ",
			TimeAgo:     "3時間前",
			Icon:        "fa-door-open",
			IconColor:   "text-green-500",
		},
		{
			Type:        "room_join",
			Title:       "【部屋参加】二つ名持ちモンスター",
			Description: "ホスト: 素材コレクター",
			TimeAgo:     "昨日",
			Icon:        "fa-right-to-bracket",
			IconColor:   "text-blue-500",
		},
		{
			Type:        "follow_add",
			Title:       "ハンター太郎さんをフォローしました",
			Description: "",
			TimeAgo:     "2日前",
			Icon:        "fa-user-plus",
			IconColor:   "text-yellow-500",
		},
	}
}

func (ph *ProfileHandler) getMockRooms() []RoomSummary {
	return []RoomSummary{
		{
			ID:          uuid.New(),
			Name:        "テスト部屋（更新済み）",
			Description: "部屋設定の更新機能が正常に動作することを確認しました。",
			GameVersion: "MHP3",
			PlayerCount: "1/4",
			Status:      "active",
			CreatedAt:   "3時間前",
		},
		{
			ID:          uuid.New(),
			Name:        "古龍種連戦",
			Description: "古龍種を順番に討伐していきます",
			GameVersion: "MHP2G",
			PlayerCount: "0/4",
			Status:      "ended",
			CreatedAt:   "1日前",
		},
	}
}

func (ph *ProfileHandler) getMockFollowers() []Follower {
	return []Follower{
		{
			ID:             uuid.New(),
			Username:       "ハンター太郎",
			AvatarURL:      "/static/images/default-avatar.png",
			IsOnline:       true,
			FollowingSince: "2日前",
		},
		{
			ID:             uuid.New(),
			Username:       "素材コレクター",
			AvatarURL:      "/static/images/default-avatar.png",
			IsOnline:       false,
			FollowingSince: "5日前",
		},
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

// getStringValue ポインタ文字列から値を安全に取得
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// formatRelativeTime 相対時間を日本語で表示
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "たった今"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d分前", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d時間前", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d日前", days)
	} else if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / (24 * 7))
		return fmt.Sprintf("%d週間前", weeks)
	} else {
		months := int(diff.Hours() / (24 * 30))
		return fmt.Sprintf("%d ヶ月前", months)
	}
}

// roomToSummary models.RoomをRoomSummaryに変換
func roomToSummary(room models.Room) RoomSummary {
	var description string
	if room.Description != nil {
		description = *room.Description
	}

	// ステータス判定
	var status, statusColor string
	var isClickable bool

	if !room.IsActive {
		status = "削除済み"
		statusColor = "text-gray-500"
		isClickable = false
	} else if room.IsClosed {
		status = "終了"
		statusColor = "text-red-600"
		isClickable = false
	} else {
		status = "アクティブ"
		statusColor = "text-green-600"
		isClickable = true
	}

	playerCount := fmt.Sprintf("%d/%d人", room.CurrentPlayers, room.MaxPlayers)

	return RoomSummary{
		ID:          room.ID,
		Name:        room.Name,
		Description: description,
		GameVersion: room.GameVersion.Code,
		PlayerCount: playerCount,
		Status:      status,
		StatusColor: statusColor,
		CreatedAt:   formatRelativeTime(room.CreatedAt),
		IsClickable: isClickable,
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func getUserFromContext(ctx context.Context) *models.User {
	// 認証ミドルウェアからユーザー情報を取得
	if user, ok := ctx.Value(middleware.DBUserContextKey).(*models.User); ok {
		return user
	}
	return nil
}
