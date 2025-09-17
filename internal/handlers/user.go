package handlers

import (
	"fmt"
	"net/http"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)


// UserHandler 他ユーザー関連のハンドラー
type UserHandler struct {
	BaseHandler
}

// NewUserHandler 新しいUserHandlerインスタンスを作成
func NewUserHandler(repo *repository.Repository) *UserHandler {
	return &UserHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
	}
}

// UserProfileData 他ユーザーのプロフィール用データ構造
type UserProfileData struct {
	User           *models.User      `json:"user"`
	IsOwnProfile   bool              `json:"isOwnProfile"`
	IsAuthenticated bool             `json:"isAuthenticated"`
	RelationStatus string            `json:"relationStatus"` // none, following, follower, mutual, blocked
	Activities     []Activity        `json:"activities"`
	Rooms          []RoomSummary     `json:"rooms"`
	Followers      []Follower        `json:"followers"`
	FollowerCount  int64             `json:"followerCount"`
	FavoriteGames  []string          `json:"favoriteGames"`
	PlayTimes      *models.PlayTimes `json:"playTimes"`
}

// Show 他のユーザーのプロフィールページを表示
func (uh *UserHandler) Show(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "uuid")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
		return
	}

	// データベースからユーザー情報を取得
	user, err := uh.repo.User.FindUserByID(userID)
	if err != nil {
		http.Error(w, "ユーザーが見つかりません", http.StatusNotFound)
		return
	}

	// 現在のユーザーと比較して自分のプロフィールかどうか判定
	currentUser := uh.getCurrentUser(r)
	isOwnProfile := false
	relationStatus := "none"
	var followerCount int64 = 0

	if currentUser != nil {
		if currentUser.ID == user.ID {
			// 自分のプロフィールの場合は/profileにリダイレクト
			http.Redirect(w, r, "/profile", http.StatusFound)
			return
		}
		// 認証済みユーザーのみフォロー関係をチェック
		relationStatus = uh.checkRelationStatus(currentUser.ID, user.ID)

		// 認証済みユーザーのみフォロワー数を取得
		if uh.repo != nil && uh.repo.UserFollow != nil {
			followers, err := uh.repo.UserFollow.GetFollowers(user.ID)
			if err == nil {
				followerCount = int64(len(followers))
			}
		}
	}

	// お気に入りゲームとプレイ時間帯を取得
	favoriteGames, _ := user.GetFavoriteGames()
	playTimes, _ := user.GetPlayTimes()

	// 実際に作成した部屋を取得
	rooms, err := uh.repo.Room.GetRoomsByHostUser(user.ID, 10, 0)
	var roomSummaries []RoomSummary
	if err == nil {
		for _, room := range rooms {
			roomSummaries = append(roomSummaries, roomToSummary(room))
		}
	}

	profileData := UserProfileData{
		User:            user,
		IsOwnProfile:    isOwnProfile,
		IsAuthenticated: currentUser != nil,
		RelationStatus:  relationStatus,
		Activities:      uh.getMockActivities(),
		Rooms:           roomSummaries,
		Followers:       uh.getMockFollowers(),
		FollowerCount:   followerCount,
		FavoriteGames:   favoriteGames,
		PlayTimes:       playTimes,
	}

	data := TemplateData{
		Title:    user.DisplayName + "のプロフィール",
		PageData: profileData,
	}

	renderTemplate(w, "user_profile.tmpl", data)
}

// GetUserProfile APIエンドポイント：ユーザープロフィール情報を取得
func (uh *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "uuid")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "無効なユーザーIDです"})
		return
	}

	user, err := uh.repo.User.FindUserByID(userID)
	if err != nil {
		respondWithJSON(w, http.StatusNotFound, map[string]string{"error": "ユーザーが見つかりません"})
		return
	}

	// 現在のユーザーと比較
	currentUser := uh.getCurrentUser(r)
	isOwnProfile := false
	relationStatus := "none"
	var followerCount int64 = 0

	if currentUser != nil {
		if currentUser.ID == user.ID {
			isOwnProfile = true
		} else {
			// 認証済みユーザーのみフォロー関係をチェック
			relationStatus = uh.checkRelationStatus(currentUser.ID, user.ID)
		}

		// 認証済みユーザーのみフォロワー数を取得
		if uh.repo != nil && uh.repo.UserFollow != nil {
			followers, err := uh.repo.UserFollow.GetFollowers(user.ID)
			if err == nil {
				followerCount = int64(len(followers))
			}
		}
	}

	// お気に入りゲームとプレイ時間帯を取得
	favoriteGames, _ := user.GetFavoriteGames()
	playTimes, _ := user.GetPlayTimes()

	// 実際に作成した部屋を取得
	rooms, err := uh.repo.Room.GetRoomsByHostUser(user.ID, 10, 0)
	var roomSummaries []RoomSummary
	if err == nil {
		for _, room := range rooms {
			roomSummaries = append(roomSummaries, roomToSummary(room))
		}
	}

	profileData := UserProfileData{
		User:            user,
		IsOwnProfile:    isOwnProfile,
		IsAuthenticated: currentUser != nil,
		RelationStatus:  relationStatus,
		Activities:      uh.getMockActivities(),
		Rooms:           roomSummaries,
		Followers:       uh.getMockFollowers(),
		FollowerCount:   followerCount,
		FavoriteGames:   favoriteGames,
		PlayTimes:       playTimes,
	}

	respondWithJSON(w, http.StatusOK, profileData)
}

// checkRelationStatus 2人のユーザー間の関係性をチェック
func (uh *UserHandler) checkRelationStatus(currentUserID, targetUserID uuid.UUID) string {
	// 相互フォローのチェック
	isMutual, err := uh.repo.UserFollow.IsMutualFollow(currentUserID, targetUserID)
	if err == nil && isMutual {
		return "mutual"
	}

	// currentUserがtargetUserをフォローしているかチェック
	follow, err := uh.repo.UserFollow.GetFollow(currentUserID, targetUserID)
	if err == nil && follow != nil && follow.Status == models.FollowStatusAccepted {
		return "following"
	}

	// targetUserがcurrentUserをフォローしているかチェック
	follow, err = uh.repo.UserFollow.GetFollow(targetUserID, currentUserID)
	if err == nil && follow != nil && follow.Status == models.FollowStatusAccepted {
		return "follower"
	}

	// TODO: ブロック機能が実装されたらここでチェック

	return "none"
}

// getCurrentUser リクエストから現在のユーザーを取得
func (uh *UserHandler) getCurrentUser(r *http.Request) *models.User {
	// 認証ミドルウェアからユーザー情報を取得
	if user, ok := r.Context().Value(middleware.DBUserContextKey).(*models.User); ok {
		return user
	}
	return nil
}

// GetProfileCard プロフィールカードの部分HTML取得
func (uh *UserHandler) GetProfileCard(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "uuid")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
		return
	}

	user, err := uh.repo.User.FindUserByID(userID)
	if err != nil {
		http.Error(w, "ユーザーが見つかりません", http.StatusNotFound)
		return
	}

	currentUser := uh.getCurrentUser(r)
	isOwnProfile := false
	relationStatus := "none"

	if currentUser != nil {
		if currentUser.ID == user.ID {
			isOwnProfile = true
		} else {
			relationStatus = uh.checkRelationStatus(currentUser.ID, user.ID)
		}
	}

	profileData := struct {
		User            *models.User
		IsOwnProfile    bool
		IsAuthenticated bool
		RelationStatus  string
		AvatarURL       string
	}{
		User:            user,
		IsOwnProfile:    isOwnProfile,
		IsAuthenticated: currentUser != nil,
		RelationStatus:  relationStatus,
		AvatarURL:       getAvatarURL(user),
	}

	if err := renderPartialTemplate(w, "profile_card_content.tmpl", profileData); err != nil {
		http.Error(w, "テンプレートエラー: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Helper functions for profile card generation
func getAvatarURL(user *models.User) string {
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		return *user.AvatarURL
	}
	return "/static/images/default-avatar.png"
}

func getBioHTML(bio *string) string {
	if bio != nil && *bio != "" {
		return fmt.Sprintf(`<p class="text-center text-gray-600 mb-6 text-sm">%s</p>`, *bio)
	}
	return ""
}


// Helper functions（一時的なモックデータ）
func (uh *UserHandler) getMockActivities() []Activity {
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
	}
}

func (uh *UserHandler) getMockFollowers() []Follower {
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
