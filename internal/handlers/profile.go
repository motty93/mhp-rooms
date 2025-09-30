package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	BaseHandler
	logger   *log.Logger
	uploader *storage.GCSUploader
	jwtAuth  *middleware.JWTAuth
}

func NewProfileHandler(repo *repository.Repository, jwtAuth *middleware.JWTAuth) *ProfileHandler {
	// GCSアップローダーを初期化
	uploader, err := storage.NewGCSUploader(context.Background())
	if err != nil {
		log.Printf("GCSアップローダーの初期化に失敗: %v", err)
		// 開発環境では警告のみ、本番では必須
		uploader = nil
	}

	return &ProfileHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		logger:   log.New(log.Writer(), "[ProfileHandler] ", log.LstdFlags),
		uploader: uploader,
		jwtAuth:  jwtAuth,
	}
}

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

type Activity struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	TimeAgo     string `json:"timeAgo"`
	Icon        string `json:"icon"`
	IconColor   string `json:"iconColor"`
}

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

type Follower struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	AvatarURL      string    `json:"avatarUrl"`
	IsOnline       bool      `json:"isOnline"`
	FollowingSince string    `json:"followingSince"`
}

func (ph *ProfileHandler) Profile(w http.ResponseWriter, r *http.Request) {
	contextUser := getUserFromContext(r.Context())
	if contextUser == nil {
		// 認証が必要 - 開発環境も本番環境も同じ処理
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	user, err := ph.repo.User.FindUserByID(contextUser.ID)
	if err != nil {
		ph.logger.Printf("ユーザー情報取得エラー: %v", err)
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	favoriteGames, _ := user.GetFavoriteGames()
	playTimes, _ := user.GetPlayTimes()

	// フォロワー数を取得（開発環境では20人固定）
	var followerCount int64 = 20
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

// EditForm プロフィール編集フォームを返す（htmx用）
func (ph *ProfileHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	contextUser := getUserFromContext(r.Context())
	if contextUser == nil {
		// 認証が必要 - 開発環境も本番環境も同じ処理
		// htmxリクエストの場合はHX-Redirectヘッダーを使用
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/auth/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// 通常のリクエストの場合はリダイレクト
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	// データベースから最新のユーザー情報を取得
	user, err := ph.repo.User.FindUserByID(contextUser.ID)
	if err != nil {
		ph.logger.Printf("ユーザー情報取得エラー: %v", err)
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// お気に入りゲームとプレイ時間帯を取得
	favoriteGames, _ := user.GetFavoriteGames()
	playTimes, _ := user.GetPlayTimes()

	// テンプレート用データ
	data := struct {
		User          *models.User
		FavoriteGames []string
		PlayTimes     *models.PlayTimes
	}{
		User:          user,
		FavoriteGames: favoriteGames,
		PlayTimes:     playTimes,
	}

	if err := renderPartialTemplate(w, "profile_edit_form", data); err != nil {
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

	userIDParam := chi.URLParam(r, "uuid")
	if userIDParam != "" {
		// 他のユーザーのプロフィール
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

	// データベースからアクティビティを取得（過去2週間分）
	userActivities, err := ph.repo.UserActivity.GetUserActivities(targetUserID, 100, 0)
	if err != nil {
		ph.logger.Printf("アクティビティ取得エラー: %v", err)
		// エラー時はフォールバック（空の配列を返す）
		userActivities = []models.UserActivity{}
	}

	// 過去2週間のアクティビティのみフィルタリング
	twoWeeksAgo := time.Now().AddDate(0, 0, -14)
	var filteredActivities []models.UserActivity
	for _, activity := range userActivities {
		if activity.CreatedAt.After(twoWeeksAgo) {
			filteredActivities = append(filteredActivities, activity)
		}
	}
	userActivities = filteredActivities

	// 最大20件に制限
	if len(userActivities) > 20 {
		userActivities = userActivities[:20]
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

	if err := renderPartialTemplate(w, "profile_activity", data); err != nil {
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
			// htmxリクエストの場合はHX-Redirectヘッダーを使用
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/auth/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// 通常のリクエストの場合はリダイレクト
			http.Redirect(w, r, "/auth/login", http.StatusFound)
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
	if err := renderPartialTemplate(w, "profile_rooms", data); err != nil {
		ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
		http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
		return
	}
}

// Following フォロー中タブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Following(w http.ResponseWriter, r *http.Request) {
	// URLパラメータからユーザーIDを取得（他ユーザーのプロフィール表示用）
	var targetUserID uuid.UUID
	var err error

	userIDParam := chi.URLParam(r, "uuid")
	if userIDParam != "" {
		// 他のユーザーのプロフィール
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

	ph.logger.Printf("フォロー中データを取得中 - ユーザーID: %s", targetUserID.String())

	if err := renderPartialTemplate(w, "profile_following", nil); err != nil {
		ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
		http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
		return
	}
}

// Followers フォロワータブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Followers(w http.ResponseWriter, r *http.Request) {
	// URLパラメータからユーザーIDを取得（他ユーザーのプロフィール表示用）
	var targetUserID uuid.UUID
	var err error

	userIDParam := chi.URLParam(r, "uuid")
	if userIDParam != "" {
		// 他のユーザーのプロフィール
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

	ph.logger.Printf("フォロワーデータを取得中 - ユーザーID: %s", targetUserID.String())

	if err := renderPartialTemplate(w, "profile_followers", nil); err != nil {
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
			AvatarURL:      "/static/images/default-avatar.webp",
			IsOnline:       true,
			FollowingSince: "2日前",
		},
		{
			ID:             uuid.New(),
			Username:       "素材コレクター",
			AvatarURL:      "/static/images/default-avatar.webp",
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

// UpdateProfileRequest プロフィール更新リクエスト
type UpdateProfileRequest struct {
	DisplayName       string   `json:"display_name"`
	Bio               string   `json:"bio"`
	PSNOnlineID       string   `json:"psn_online_id"`
	NintendoNetworkID string   `json:"nintendo_network_id"`
	NintendoSwitchID  string   `json:"nintendo_switch_id"`
	TwitterID         string   `json:"twitter_id"`
	FavoriteGames     []string `json:"favorite_games"`
	PlayTimes         struct {
		Weekday string `json:"weekday"`
		Weekend string `json:"weekend"`
	} `json:"play_times"`
}

// UpdateProfile プロフィール更新APIハンドラー
func (ph *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		respondWithError(w, http.StatusUnauthorized, "認証が必要です")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ph.logger.Printf("JSONデコードエラー: %v", err)
		respondWithError(w, http.StatusBadRequest, "リクエストの形式が正しくありません")
		return
	}

	// ユーザー情報を更新
	user.DisplayName = req.DisplayName
	if req.Bio != "" {
		user.Bio = strPtr(req.Bio)
	} else {
		user.Bio = nil
	}

	// プラットフォームID
	if req.PSNOnlineID != "" {
		user.PSNOnlineID = strPtr(req.PSNOnlineID)
	} else {
		user.PSNOnlineID = nil
	}
	if req.NintendoNetworkID != "" {
		user.NintendoNetworkID = strPtr(req.NintendoNetworkID)
	} else {
		user.NintendoNetworkID = nil
	}
	if req.NintendoSwitchID != "" {
		user.NintendoSwitchID = strPtr(req.NintendoSwitchID)
	} else {
		user.NintendoSwitchID = nil
	}
	if req.TwitterID != "" {
		user.TwitterID = strPtr(req.TwitterID)
	} else {
		user.TwitterID = nil
	}

	// お気に入りゲーム
	if err := user.SetFavoriteGames(req.FavoriteGames); err != nil {
		ph.logger.Printf("お気に入りゲーム設定エラー: %v", err)
		respondWithError(w, http.StatusInternalServerError, "お気に入りゲームの設定に失敗しました")
		return
	}

	// プレイ時間帯
	playTimes := models.PlayTimes{
		Weekday: req.PlayTimes.Weekday,
		Weekend: req.PlayTimes.Weekend,
	}
	if err := user.SetPlayTimes(&playTimes); err != nil {
		ph.logger.Printf("プレイ時間帯設定エラー: %v", err)
		respondWithError(w, http.StatusInternalServerError, "プレイ時間帯の設定に失敗しました")
		return
	}

	// データベースを更新
	if err := ph.repo.User.UpdateUser(user); err != nil {
		ph.logger.Printf("ユーザー更新エラー: %v", err)
		respondWithError(w, http.StatusInternalServerError, "プロフィールの更新に失敗しました")
		return
	}

	// 成功レスポンス
	response := map[string]interface{}{
		"message": "プロフィールを更新しました",
		"user":    user,
	}
	respondWithJSON(w, http.StatusOK, response)
}

// ViewProfile プロフィールカード表示用APIハンドラー（htmx用）
func (ph *ProfileHandler) ViewProfile(w http.ResponseWriter, r *http.Request) {
	contextUser := getUserFromContext(r.Context())
	if contextUser == nil {
		// htmxリクエストの場合はHX-Redirectヘッダーを使用
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/auth/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	// データベースから最新のユーザー情報を取得
	user, err := ph.repo.User.FindUserByID(contextUser.ID)
	if err != nil {
		ph.logger.Printf("ユーザー情報取得エラー: %v", err)
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
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

	// テンプレート用データ
	data := struct {
		User          *models.User
		FavoriteGames []string
		PlayTimes     *models.PlayTimes
		FollowerCount int64
	}{
		User:          user,
		FavoriteGames: favoriteGames,
		PlayTimes:     playTimes,
		FollowerCount: followerCount,
	}

	if err := renderPartialTemplate(w, "profile_view", data); err != nil {
		ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
		http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
		return
	}
}

func (ph *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		respondWithError(w, http.StatusUnauthorized, "認証が必要です")
		return
	}

	if ph.uploader == nil {
		respondWithError(w, http.StatusServiceUnavailable, "画像アップロードサービスが利用できません")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		ph.logger.Printf("マルチパートフォーム解析エラー: %v", err)
		respondWithError(w, http.StatusBadRequest, "ファイルの解析に失敗しました")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		ph.logger.Printf("ファイル取得エラー: %v", err)
		respondWithError(w, http.StatusBadRequest, "アバター画像が選択されていません")
		return
	}
	defer file.Close()

	result, err := ph.uploader.UploadAvatar(r.Context(), user.ID.String(), file, header)
	if err != nil {
		ph.logger.Printf("アバターアップロードエラー: %v", err)
		// エラーメッセージはクライアントに詳細を返さない
		respondWithError(w, http.StatusInternalServerError, "画像のアップロードに失敗しました")
		return
	}

	user.AvatarURL = &result.URL
	if err := ph.repo.User.UpdateUser(user); err != nil {
		ph.logger.Printf("ユーザー更新エラー: %v", err)
		respondWithError(w, http.StatusInternalServerError, "プロフィール情報の更新に失敗しました")
		return
	}

	// キャッシュをクリア（jwtAuthが設定されている場合）
	if ph.jwtAuth != nil && ph.jwtAuth.GetUserCache() != nil {
		ph.logger.Printf("ユーザーキャッシュをクリア: %s", user.ID)
		ph.jwtAuth.GetUserCache().Delete(user.SupabaseUserID)
	}

	response := map[string]interface{}{
		"message":     "アバター画像を更新しました",
		"avatar_url":  result.URL,
		"object_path": result.ObjectPath,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func getUserFromContext(ctx context.Context) *models.User {
	// 認証ミドルウェアからユーザー情報を取得
	if user, ok := ctx.Value(middleware.DBUserContextKey).(*models.User); ok {
		return user
	}
	return nil
}
