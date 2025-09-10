package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"

	"github.com/google/uuid"
)

type AuthHandler struct {
	BaseHandler
	authMiddleware *middleware.JWTAuth
}

func NewAuthHandler(repo *repository.Repository) *AuthHandler {
	return &AuthHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
	}
}

// SetAuthMiddleware は認証ミドルウェアを設定
func (h *AuthHandler) SetAuthMiddleware(auth *middleware.JWTAuth) {
	h.authMiddleware = auth
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "ログイン",
	}
	renderTemplate(w, "login.tmpl", data)
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "新規登録",
	}
	renderTemplate(w, "register.tmpl", data)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

func (h *AuthHandler) PasswordResetPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "パスワードリセット",
	}
	renderTemplate(w, "password_reset.tmpl", data)
}

func (h *AuthHandler) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

func (h *AuthHandler) PasswordResetConfirmPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "パスワードリセット確認",
	}
	renderTemplate(w, "password_reset_confirm.tmpl", data)
}

func (h *AuthHandler) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

func (h *AuthHandler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

func (h *AuthHandler) AuthCallback(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "認証中",
	}
	renderTemplate(w, "auth-callback.tmpl", data)
}

func (h *AuthHandler) CurrentUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

func (h *AuthHandler) CompleteProfilePage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "プロフィール設定",
	}
	renderTemplate(w, "complete-profile.tmpl", data)
}

func (h *AuthHandler) CompleteProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "このエンドポイントは使用されません。フロントエンド認証をご利用ください。",
	})
}

// SyncUser はSupabase認証後にアプリケーションDBにユーザーを同期する
func (h *AuthHandler) SyncUser(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	dbUser, hasDBUser := middleware.GetDBUserFromContext(r.Context())

	var req struct {
		PSNId string `json:"psn_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの形式が正しくありません", http.StatusBadRequest)
		return
	}

	if psnId, ok := user.Metadata["psn_id"].(string); ok && psnId != "" {
		req.PSNId = psnId
	}

	now := time.Now()

	if !hasDBUser || dbUser == nil {
		if h.authMiddleware != nil {
			if req.PSNId != "" {
				if user.Metadata == nil {
					user.Metadata = make(map[string]interface{})
				}
				user.Metadata["psn_id"] = req.PSNId
			}

			newUser, err := h.authMiddleware.EnsureUserExistsWithContext(r.Context(), user)
			if err != nil {
				http.Error(w, "ユーザーの作成に失敗しました", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "ユーザーが正常に作成されました",
				"user": map[string]interface{}{
					"id":     newUser.ID,
					"email":  newUser.Email,
					"psn_id": newUser.PSNOnlineID,
				},
			})
			return
		}

		supabaseUserID, err := uuid.Parse(user.ID)
		if err != nil {
			http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
			return
		}

		username := user.Email
		if idx := strings.Index(user.Email, "@"); idx > 0 {
			username = user.Email[:idx]
		}
		displayName := ""

		var psnOnlineID *string
		if req.PSNId != "" {
			psnOnlineID = &req.PSNId
		}

		newUser := &models.User{
			SupabaseUserID: supabaseUserID,
			Email:          user.Email,
			Username:       &username,
			DisplayName:    displayName,
			PSNOnlineID:    psnOnlineID,
			IsActive:       true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := h.repo.User.CreateUser(newUser); err != nil {
			http.Error(w, "ユーザーの作成に失敗しました", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "ユーザーが正常に作成されました",
			"user": map[string]interface{}{
				"id":     newUser.ID,
				"email":  newUser.Email,
				"psn_id": newUser.PSNOnlineID,
			},
		})
		return
	}

	// ユーザー情報に変更がある場合のみ更新
	needsUpdate := false

	if dbUser.Email != user.Email {
		dbUser.Email = user.Email
		needsUpdate = true
	}

	if req.PSNId != "" && (dbUser.PSNOnlineID == nil || *dbUser.PSNOnlineID != req.PSNId) {
		dbUser.PSNOnlineID = &req.PSNId
		needsUpdate = true
	}

	// 更新が必要な場合のみDBを更新
	if needsUpdate {
		dbUser.UpdatedAt = now
		if err := h.repo.User.UpdateUser(dbUser); err != nil {
			http.Error(w, "ユーザー情報の更新に失敗しました", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "ユーザー情報が正常に更新されました",
		"user": map[string]interface{}{
			"id":     dbUser.ID,
			"email":  dbUser.Email,
			"psn_id": dbUser.PSNOnlineID,
		},
	})
}

// UpdatePSNId はユーザーのPSN IDを更新する
func (h *AuthHandler) UpdatePSNId(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	dbUser, hasDBUser := middleware.GetDBUserFromContext(r.Context())

	var req struct {
		PSNId string `json:"psn_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの形式が正しくありません", http.StatusBadRequest)
		return
	}

	if req.PSNId == "" {
		http.Error(w, "PSN IDは必須です", http.StatusBadRequest)
		return
	}

	if !hasDBUser || dbUser == nil {
		if h.authMiddleware != nil {
			if user.Metadata == nil {
				user.Metadata = make(map[string]interface{})
			}
			user.Metadata["psn_id"] = req.PSNId

			newUser, err := h.authMiddleware.EnsureUserExistsWithContext(r.Context(), user)
			if err != nil {
				http.Error(w, "ユーザーの作成に失敗しました", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "PSN IDが正常に設定されました",
				"psn_id":  newUser.PSNOnlineID,
			})
			return
		}

		supabaseUserID, err := uuid.Parse(user.ID)
		if err != nil {
			http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
			return
		}

		username := user.Email
		if idx := strings.Index(user.Email, "@"); idx > 0 {
			username = user.Email[:idx]
		}
		displayName := ""

		now := time.Now()
		newUser := &models.User{
			SupabaseUserID: supabaseUserID,
			Email:          user.Email,
			Username:       &username,
			DisplayName:    displayName,
			PSNOnlineID:    &req.PSNId,
			IsActive:       true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := h.repo.User.CreateUser(newUser); err != nil {
			http.Error(w, "ユーザーの作成に失敗しました", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "PSN IDが正常に設定されました",
			"psn_id":  newUser.PSNOnlineID,
		})
		return
	}

	dbUser.PSNOnlineID = &req.PSNId
	dbUser.UpdatedAt = time.Now()

	if err := h.repo.User.UpdateUser(dbUser); err != nil {
		http.Error(w, "PSN IDの更新に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "PSN IDが正常に更新されました",
		"psn_id":  dbUser.PSNOnlineID,
	})
}

// GetCurrentUser は現在ログイン中のユーザー情報を取得する
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	dbUser, hasDBUser := middleware.GetDBUserFromContext(r.Context())
	if !hasDBUser || dbUser == nil {
		http.Error(w, "ユーザー情報が見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": dbUser,
	})
}
