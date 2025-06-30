package handlers

import (
	"encoding/json"
	"net/http"

	"mhp-rooms/internal/repository"
)

type AuthHandler struct {
	BaseHandler
}

func NewAuthHandler(repo *repository.Repository) *AuthHandler {
	return &AuthHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
	}
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

// 以下のメソッドはフロントエンド認証（Supabase-js）に移行したため無効化
// 認証処理はブラウザで直接Supabaseに接続して行われます

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

// パスワードリセット関連も同様に無効化

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

// Google認証も無効化（削除済みファイルから移植）

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