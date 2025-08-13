package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	CreatedAt   string    `json:"createdAt"`
}

// Follower フォロワー情報
type Follower struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	AvatarURL     string    `json:"avatarUrl"`
	IsOnline      bool      `json:"isOnline"`
	FollowingSince string   `json:"followingSince"`
}

// Profile 自分のプロフィールページを表示
func (ph *ProfileHandler) Profile(w http.ResponseWriter, r *http.Request) {
	// 認証情報から自分のユーザー情報を取得
	user := getUserFromContext(r.Context())
	if user == nil {
		// 開発環境: 認証がない場合はテストユーザーを使用
		if os.Getenv("ENV") != "production" {
			devUser, err := ph.repo.User.FindUserByEmail("hunter1@example.com")
			if err == nil && devUser != nil {
				user = devUser
			} else {
				http.Redirect(w, r, "/auth/login", http.StatusFound)
				return
			}
		} else {
			// 認証されていない場合はログインページへリダイレクト
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}
	}

	// お気に入りゲームとプレイ時間帯を取得
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

	profileData := ProfileData{
		User:          user,
		IsOwnProfile:  true,
		Activities:    ph.getMockActivities(),
		Rooms:         ph.getMockRooms(),
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

	profileData := ProfileData{
		User:          user,
		IsOwnProfile:  isOwnProfile,
		Activities:    ph.getMockActivities(),
		Rooms:         ph.getMockRooms(),
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

	profileData := ProfileData{
		User:         user,
		IsOwnProfile: isOwnProfile,
		Activities:   ph.getMockActivities(),
		Rooms:        ph.getMockRooms(),
		Followers:    ph.getMockFollowers(),
	}

	respondWithJSON(w, http.StatusOK, profileData)
}

// EditForm プロフィール編集フォームを返す（htmx用）
func (ph *ProfileHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		user = &models.User{
			Username:    strPtr("rdwbocungelt5"),
			DisplayName: "rdwbocungelt5",
			AvatarURL:   strPtr("/static/images/default-avatar.png"),
			Bio:         strPtr("モンハン大好きです！一緒に楽しく狩りに行けるフレンドを募集しています。VCも可能です。よろしくお願いします！"),
			CreatedAt:   time.Now().AddDate(-1, -2, -15), // 1年2ヶ月15日前に登録したことにする
		}
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	bio := ""
	if user.Bio != nil {
		bio = *user.Bio
	}

	html := `
	<div class="p-6">
		<h3 class="text-xl font-bold text-gray-800 mb-4">プロフィール編集</h3>
		<div class="space-y-4">
			<div>
				<label class="text-sm font-bold text-gray-600 block mb-1">ユーザー名</label>
				<input type="text" value="` + username + `" class="w-full bg-gray-100 text-gray-800 rounded-lg p-2 border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500">
			</div>
			<div>
				<label class="text-sm font-bold text-gray-600 block mb-1">自己紹介</label>
				<textarea class="w-full bg-gray-100 text-gray-800 rounded-lg p-2 border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500" rows="4">` + bio + `</textarea>
			</div>
			<div class="flex space-x-2">
				<button class="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg transition-colors">保存</button>
				<button hx-get="/api/profile/view" hx-target="#profile-card" hx-swap="outerHTML" class="w-full bg-gray-200 hover:bg-gray-300 text-gray-800 font-bold py-2 px-4 rounded-lg transition-colors">キャンセル</button>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Activity アクティビティタブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Activity(w http.ResponseWriter, r *http.Request) {
	activities := ph.getMockActivities()
	
	html := `<div><h3 class="text-xl font-bold mb-4 text-gray-800">最近の活動履歴</h3><div class="space-y-4">`
	
	for _, activity := range activities {
		html += `
		<div class="bg-gray-50 p-4 rounded-lg flex justify-between items-center hover:bg-gray-100 transition-colors border border-gray-200">
			<div class="flex items-center space-x-4">
				<i class="fa-solid ` + activity.Icon + ` ` + activity.IconColor + `"></i>
				<div>
					<p class="font-semibold text-gray-800">` + activity.Title + `</p>
					<p class="text-sm text-gray-500">` + activity.Description + `</p>
				</div>
			</div>
			<span class="text-xs text-gray-400">` + activity.TimeAgo + `</span>
		</div>`
	}
	
	html += `</div></div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Rooms 作成した部屋タブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Rooms(w http.ResponseWriter, r *http.Request) {
	html := `
	<div>
		<h3 class="text-xl font-bold mb-4 text-gray-800">作成した部屋</h3>
		<div class="space-y-4">
			<div class="bg-gray-50 p-4 rounded-lg border border-gray-200">
				<div class="flex justify-between items-start mb-2">
					<h4 class="font-semibold text-gray-800">テスト部屋（更新済み）</h4>
					<span class="text-xs text-gray-500">3時間前</span>
				</div>
				<p class="text-sm text-gray-600 mb-2">部屋設定の更新機能が正常に動作することを確認しました。</p>
				<div class="flex items-center space-x-4 text-xs text-gray-500">
					<span>MHP3</span>
					<span>1/4人</span>
					<span class="text-green-600">アクティブ</span>
				</div>
			</div>
			<div class="bg-gray-50 p-4 rounded-lg border border-gray-200">
				<div class="flex justify-between items-start mb-2">
					<h4 class="font-semibold text-gray-800">古龍種連戦</h4>
					<span class="text-xs text-gray-500">1日前</span>
				</div>
				<p class="text-sm text-gray-600 mb-2">古龍種を順番に討伐していきます</p>
				<div class="flex items-center space-x-4 text-xs text-gray-500">
					<span>MHP2G</span>
					<span>0/4人</span>
					<span class="text-red-600">終了</span>
				</div>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Following フォロー中タブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Following(w http.ResponseWriter, r *http.Request) {
	html := `
	<div>
		<h3 class="text-xl font-bold mb-4 text-gray-800">フォロー中のユーザー</h3>
		<div class="space-y-4">
			<div class="bg-gray-50 p-4 rounded-lg flex items-center justify-between border border-gray-200">
				<div class="flex items-center space-x-4">
					<img class="w-10 h-10 rounded-full object-cover" src="/static/images/default-avatar.png" alt="古龍ハンター">
					<div>
						<p class="font-semibold text-gray-800">古龍ハンター</p>
						<p class="text-xs text-gray-500">オンライン</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span class="w-3 h-3 bg-green-500 rounded-full"></span>
					<span class="text-xs text-gray-500">3日前からフォロー中</span>
				</div>
			</div>
			<div class="bg-gray-50 p-4 rounded-lg flex items-center justify-between border border-gray-200">
				<div class="flex items-center space-x-4">
					<img class="w-10 h-10 rounded-full object-cover" src="/static/images/default-avatar.png" alt="モンス討伐王">
					<div>
						<p class="font-semibold text-gray-800">モンス討伐王</p>
						<p class="text-xs text-gray-500">オフライン</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span class="w-3 h-3 bg-gray-400 rounded-full"></span>
					<span class="text-xs text-gray-500">1週間前からフォロー中</span>
				</div>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Followers フォロワータブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Followers(w http.ResponseWriter, r *http.Request) {
	html := `
	<div>
		<h3 class="text-xl font-bold mb-4 text-gray-800">フォロワーリスト</h3>
		<div class="space-y-4">
			<div class="bg-gray-50 p-4 rounded-lg flex items-center justify-between border border-gray-200">
				<div class="flex items-center space-x-4">
					<img class="w-10 h-10 rounded-full object-cover" src="/static/images/default-avatar.png" alt="ハンター太郎">
					<div>
						<p class="font-semibold text-gray-800">ハンター太郎</p>
						<p class="text-xs text-gray-500">オンライン</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span class="w-3 h-3 bg-green-500 rounded-full"></span>
					<span class="text-xs text-gray-500">2日前からフォロー中</span>
				</div>
			</div>
			<div class="bg-gray-50 p-4 rounded-lg flex items-center justify-between border border-gray-200">
				<div class="flex items-center space-x-4">
					<img class="w-10 h-10 rounded-full object-cover" src="/static/images/default-avatar.png" alt="素材コレクター">
					<div>
						<p class="font-semibold text-gray-800">素材コレクター</p>
						<p class="text-xs text-gray-500">オフライン</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span class="w-3 h-3 bg-gray-400 rounded-full"></span>
					<span class="text-xs text-gray-500">5日前からフォロー中</span>
				</div>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
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
			ID:            uuid.New(),
			Username:      "ハンター太郎",
			AvatarURL:     "/static/images/default-avatar.png",
			IsOnline:      true,
			FollowingSince: "2日前",
		},
		{
			ID:            uuid.New(),
			Username:      "素材コレクター",
			AvatarURL:     "/static/images/default-avatar.png",
			IsOnline:      false,
			FollowingSince: "5日前",
		},
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
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