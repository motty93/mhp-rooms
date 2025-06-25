package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ProfileCompleteMiddleware プロフィール完成チェックミドルウェア
func (h *Handler) ProfileCompleteMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 認証状態を確認
		userID := r.Context().Value("user_id")
		if userID == nil {
			// 未認証の場合はそのまま通す
			next.ServeHTTP(w, r)
			return
		}

		// プロフィール補完関連のパスは除外
		if r.URL.Path == "/auth/complete-profile" || 
		   r.URL.Path == "/api/user/current" ||
		   r.URL.Path == "/auth/logout" {
			next.ServeHTTP(w, r)
			return
		}

		// ユーザー情報を取得
		user, err := h.repo.FindUserByID(userID.(uuid.UUID))
		if err != nil {
			// エラーの場合はログインページへ
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		// PSN IDが未設定の場合はプロフィール補完ページへ
		if user.PSNOnlineID == nil || *user.PSNOnlineID == "" {
			// API呼び出しの場合はエラーレスポンス
			if r.Header.Get("Content-Type") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPreconditionRequired)
				w.Write([]byte(`{"message":"プロフィールを完成させてください","redirect":"/auth/complete-profile"}`))
				return
			}
			// 通常のリクエストはリダイレクト
			http.Redirect(w, r, "/auth/complete-profile", http.StatusTemporaryRedirect)
			return
		}

		// プロフィールが完成している場合は次のハンドラーへ
		next.ServeHTTP(w, r)
	}
}

// AuthMiddleware 認証チェックミドルウェア（例）
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// セッションまたはJWTから認証情報を取得
		// この実装は簡略化されています
		
		// 仮実装: セッションからuser_idを取得
		// 実際の実装では、セッションストアやJWT検証を行う
		userIDStr := r.Header.Get("X-User-ID") // 仮のヘッダー
		if userIDStr == "" {
			// 未認証
			next.ServeHTTP(w, r)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			// 無効なユーザーID
			next.ServeHTTP(w, r)
			return
		}

		// コンテキストにユーザーIDを設定
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireAuthMiddleware 認証必須ミドルウェア
func (h *Handler) RequireAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID == nil {
			// API呼び出しの場合
			if r.Header.Get("Content-Type") == "application/json" ||
			   r.Header.Get("Accept") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message":"認証が必要です"}`))
				return
			}
			// 通常のリクエストはログインページへリダイレクト
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	}
}