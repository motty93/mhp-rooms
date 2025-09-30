package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/services"
)

type FollowHandler struct {
	BaseHandler
	activityService *services.ActivityService
	logger          *log.Logger
}

func NewFollowHandler(repo *repository.Repository) *FollowHandler {
	return &FollowHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		activityService: services.NewActivityService(repo),
		logger:          log.New(log.Writer(), "[FollowHandler] ", log.LstdFlags),
	}
}

// FollowUser ユーザーをフォローする
func (fh *FollowHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていません", http.StatusUnauthorized)
		return
	}

	followerUserID := dbUser.ID

	// フォロー対象のユーザーIDを取得
	followingUserIDStr := chi.URLParam(r, "userID")
	followingUserID, err := uuid.Parse(followingUserIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
		return
	}

	// 自分自身をフォローすることはできない
	if followerUserID == followingUserID {
		http.Error(w, "自分自身をフォローすることはできません", http.StatusBadRequest)
		return
	}

	// フォロー対象のユーザーが存在するかチェック
	followingUser, err := fh.repo.User.FindUserByID(followingUserID)
	if err != nil {
		fh.logger.Printf("フォロー対象ユーザーの取得エラー: %v", err)
		http.Error(w, "フォロー対象のユーザーが見つかりません", http.StatusNotFound)
		return
	}

	// 既にフォロー関係があるかチェック
	existingFollow, err := fh.repo.UserFollow.GetFollow(followerUserID, followingUserID)
	if err == nil && existingFollow != nil {
		// 既にフォローしている場合でもプロフィールカードのHTMLを返す
		fh.logger.Printf("既存のフォロー関係が見つかりました: follower=%s, following=%s", followerUserID, followingUserID)
		fh.returnProfileCardHTML(w, r, followingUser, dbUser)
		return
	}
	if err != nil {
		fh.logger.Printf("GetFollow エラー: %v", err)
	}

	// フォロー関係を作成
	userFollow := &models.UserFollow{
		BaseModel:       models.BaseModel{ID: uuid.New()},
		FollowerUserID:  followerUserID,
		FollowingUserID: followingUserID,
		Status:          "accepted", // 現在の実装では自動承認
	}

	if err := fh.repo.UserFollow.CreateFollow(userFollow); err != nil {
		fh.logger.Printf("フォロー作成エラー: %v", err)
		http.Error(w, "フォロー処理に失敗しました", http.StatusInternalServerError)
		return
	}

	// アクティビティを記録（失敗してもメイン処理は続行）
	if err := fh.activityService.RecordFollow(followerUserID, followingUserID, followingUser); err != nil {
		fh.logger.Printf("フォローアクティビティの記録に失敗: %v", err)
		// アクティビティ記録失敗はメイン処理に影響させない
	}

	// プロフィールカードのHTMLを返す
	fh.returnProfileCardHTML(w, r, followingUser, dbUser)
}

// UnfollowUser ユーザーのフォローを解除する
func (fh *FollowHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	fh.logger.Printf("UnfollowUser ハンドラーが呼び出されました")

	// 認証チェック
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		fh.logger.Printf("認証エラー: ユーザーが見つかりません")
		http.Error(w, "認証されていません", http.StatusUnauthorized)
		return
	}

	followerUserID := dbUser.ID
	fh.logger.Printf("認証ユーザーID: %s", followerUserID)

	// フォロー解除対象のユーザーIDを取得
	followingUserIDStr := chi.URLParam(r, "userID")
	followingUserID, err := uuid.Parse(followingUserIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
		return
	}

	// デバッグログ
	fh.logger.Printf("フォロー解除: followerUserID=%s, followingUserID=%s", followerUserID, followingUserID)

	// フォロー関係の存在チェック
	existingFollow, err := fh.repo.UserFollow.GetFollow(followerUserID, followingUserID)
	if err != nil {
		fh.logger.Printf("フォロー関係の取得エラー: %v", err)
		http.Error(w, "フォロー解除処理に失敗しました", http.StatusInternalServerError)
		return
	}
	if existingFollow == nil {
		fh.logger.Printf("フォロー関係が見つかりません: follower=%s, following=%s", followerUserID, followingUserID)
		http.Error(w, "フォロー関係が見つかりません", http.StatusNotFound)
		return
	}

	// フォロー関係を削除
	if err := fh.repo.UserFollow.DeleteFollow(followerUserID, followingUserID); err != nil {
		fh.logger.Printf("フォロー削除エラー: %v", err)
		http.Error(w, "フォロー解除に失敗しました", http.StatusInternalServerError)
		return
	}

	// フォロー対象のユーザー情報を取得（アクティビティ記録用）
	followingUser, userErr := fh.repo.User.FindUserByID(followingUserID)

	// フォロー解除後のプロフィールカードのHTMLを返す
	if userErr == nil && followingUser != nil {
		fh.returnProfileCardHTML(w, r, followingUser, dbUser)
	} else {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
	}
}

// GetFollowStatus フォロー状態を取得する
func (fh *FollowHandler) GetFollowStatus(w http.ResponseWriter, r *http.Request) {
	// 認証チェック
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていません", http.StatusUnauthorized)
		return
	}

	followerUserID := dbUser.ID

	// ターゲットユーザーIDを取得
	targetUserIDStr := chi.URLParam(r, "userID")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
		return
	}

	// 自分自身との関係は常にfalse
	if followerUserID == targetUserID {
		response := map[string]interface{}{
			"is_following":     false,
			"is_followed_by":   false,
			"is_mutual_follow": false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// フォロー関係をチェック
	isFollowing := false
	followRelation, err := fh.repo.UserFollow.GetFollow(followerUserID, targetUserID)
	if err == nil && followRelation != nil && followRelation.Status == "accepted" {
		isFollowing = true
	}

	// 逆方向のフォロー関係をチェック
	isFollowedBy := false
	reverseFollowRelation, err := fh.repo.UserFollow.GetFollow(targetUserID, followerUserID)
	if err == nil && reverseFollowRelation != nil && reverseFollowRelation.Status == "accepted" {
		isFollowedBy = true
	}

	// 相互フォローかどうか
	isMutualFollow := isFollowing && isFollowedBy

	response := map[string]interface{}{
		"is_following":     isFollowing,
		"is_followed_by":   isFollowedBy,
		"is_mutual_follow": isMutualFollow,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// returnProfileCardHTML プロフィールカードのHTMLを返す
func (fh *FollowHandler) returnProfileCardHTML(w http.ResponseWriter, r *http.Request, targetUser *models.User, currentUser *models.User) {
	// フォロー関係をチェック
	relationStatus := fh.checkRelationStatus(currentUser.ID, targetUser.ID)

	profileData := struct {
		User            *models.User
		IsOwnProfile    bool
		IsAuthenticated bool
		RelationStatus  string
		AvatarURL       string
	}{
		User:            targetUser,
		IsOwnProfile:    false,
		IsAuthenticated: currentUser != nil,
		RelationStatus:  relationStatus,
		AvatarURL:       fh.getAvatarURL(targetUser),
	}

	if err := renderPartialTemplate(w, "profile_card_content", profileData); err != nil {
		http.Error(w, "テンプレートエラー: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// checkRelationStatus 2人のユーザー間の関係性をチェック（user.goと同じロジック）
func (fh *FollowHandler) checkRelationStatus(currentUserID, targetUserID uuid.UUID) string {
	// 相互フォローのチェック
	isMutual, err := fh.repo.UserFollow.IsMutualFollow(currentUserID, targetUserID)
	if err == nil && isMutual {
		return "mutual"
	}

	// currentUserがtargetUserをフォローしているかチェック
	follow, err := fh.repo.UserFollow.GetFollow(currentUserID, targetUserID)
	if err == nil && follow != nil && follow.Status == models.FollowStatusAccepted {
		return "following"
	}

	// targetUserがcurrentUserをフォローしているかチェック
	follow, err = fh.repo.UserFollow.GetFollow(targetUserID, currentUserID)
	if err == nil && follow != nil && follow.Status == models.FollowStatusAccepted {
		return "follower"
	}

	return "none"
}

// ヘルパー関数
func (fh *FollowHandler) getAvatarURL(user *models.User) string {
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		return *user.AvatarURL
	}
	return "/static/images/default-avatar.webp"
}
