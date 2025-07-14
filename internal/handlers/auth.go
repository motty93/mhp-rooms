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
	renderTemplate(w, "login.html", data)
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "新規登録",
	}
	renderTemplate(w, "register.html", data)
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
	renderTemplate(w, "password_reset.html", data)
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
	renderTemplate(w, "password_reset_confirm.html", data)
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
	renderTemplate(w, "auth-callback.html", data)
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
	renderTemplate(w, "complete-profile.html", data)
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

	// コンテキストからDBユーザー情報を取得
	dbUser, hasDBUser := middleware.GetDBUserFromContext(r.Context())

	var req struct {
		PSNId string `json:"psn_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの形式が正しくありません", http.StatusBadRequest)
		return
	}

	// PSN IDの優先度: メタデータ > リクエスト
	if psnId, ok := user.Metadata["psn_id"].(string); ok && psnId != "" {
		req.PSNId = psnId
	}

	now := time.Now()

	// DBユーザーが存在しない場合は新規作成
	if !hasDBUser || dbUser == nil {
		// ミドルウェアのメソッドを使用してユーザーを作成
		if h.authMiddleware != nil {
			// PSN IDをメタデータに追加
			if req.PSNId != "" {
				if user.Metadata == nil {
					user.Metadata = make(map[string]interface{})
				}
				user.Metadata["psn_id"] = req.PSNId
			}

			newUser, err := h.authMiddleware.EnsureUserExists(user)
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

		// フォールバック: ミドルウェアがない場合は直接作成
		supabaseUserID, err := uuid.Parse(user.ID)
		if err != nil {
			http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
			return
		}

		displayName := user.Email
		if idx := strings.Index(user.Email, "@"); idx > 0 {
			displayName = user.Email[:idx]
		}

		var psnOnlineID *string
		if req.PSNId != "" {
			psnOnlineID = &req.PSNId
		}

		newUser := &models.User{
			SupabaseUserID: supabaseUserID,
			Email:          user.Email,
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

	// 既存ユーザーの情報を更新
	dbUser.Email = user.Email
	if req.PSNId != "" {
		dbUser.PSNOnlineID = &req.PSNId
	}
	dbUser.UpdatedAt = now

	if err := h.repo.User.UpdateUser(dbUser); err != nil {
		http.Error(w, "ユーザー情報の更新に失敗しました", http.StatusInternalServerError)
		return
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

	// コンテキストからDBユーザー情報を取得
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

	// DBユーザーが存在しない場合は新規作成
	if !hasDBUser || dbUser == nil {
		// ミドルウェアのメソッドを使用してユーザーを作成
		if h.authMiddleware != nil {
			// PSN IDをメタデータに追加
			if user.Metadata == nil {
				user.Metadata = make(map[string]interface{})
			}
			user.Metadata["psn_id"] = req.PSNId

			newUser, err := h.authMiddleware.EnsureUserExists(user)
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

		// フォールバック: ミドルウェアがない場合は直接作成
		supabaseUserID, err := uuid.Parse(user.ID)
		if err != nil {
			http.Error(w, "無効なユーザーIDです", http.StatusBadRequest)
			return
		}

		displayName := user.Email
		if idx := strings.Index(user.Email, "@"); idx > 0 {
			displayName = user.Email[:idx]
		}

		now := time.Now()
		newUser := &models.User{
			SupabaseUserID: supabaseUserID,
			Email:          user.Email,
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

	// 既存ユーザーのPSN IDを更新
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
