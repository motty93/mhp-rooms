package handlers

import (
	"context"
	"net/http"
	"strings"

	"mhp-rooms/internal/models"

	"github.com/google/uuid"
)

func (h *BaseHandler) ProfileCompleteMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID == nil {
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/auth/complete-profile" ||
			r.URL.Path == "/api/user/current" ||
			r.URL.Path == "/auth/logout" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := h.repo.FindUserByID(userID.(uuid.UUID))
		if err != nil {
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}

		if user.PSNOnlineID == nil || *user.PSNOnlineID == "" {
			if r.Header.Get("Content-Type") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPreconditionRequired)
				w.Write([]byte(`{"message":"プロフィールを完成させてください","redirect":"/auth/complete-profile"}`))
				return
			}
			http.Redirect(w, r, "/auth/complete-profile", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	}
}

type contextKey string

const (
	userContextKey  = contextKey("user")
	tokenContextKey = contextKey("token")
)

func (h *BaseHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var accessToken string

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			accessToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			cookie, err := r.Cookie("sb-access-token")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			accessToken = cookie.Value
		}

		authClient := h.supabase.Auth.WithToken(accessToken)

		userResp, err := authClient.GetUser()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user := userResp.User

		supabaseUserID, err := uuid.Parse(user.ID.String())
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		dbUser, err := h.repo.FindUserBySupabaseUserID(supabaseUserID)
		if err != nil || dbUser == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, dbUser)
		ctx = context.WithValue(ctx, tokenContextKey, accessToken)
		ctx = context.WithValue(ctx, "user_id", dbUser.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}

func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenContextKey).(string)
	return token, ok
}

func (h *BaseHandler) RequireAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID == nil {
			if r.Header.Get("Content-Type") == "application/json" ||
				r.Header.Get("Accept") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message":"認証が必要です"}`))
				return
			}
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	}
}
