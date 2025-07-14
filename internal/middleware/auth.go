package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserContextKey   contextKey = "user"
	DBUserContextKey contextKey = "dbUser"
)

type SupabaseJWTClaims struct {
	jwt.RegisteredClaims
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone"`
	AppMetadata  map[string]interface{} `json:"app_metadata"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
}

type AuthUser struct {
	ID       string
	Email    string
	Metadata map[string]interface{}
}

type JWTAuth struct {
	jwtSecret []byte
	repo      *repository.Repository
}

func NewJWTAuth(repo *repository.Repository) (*JWTAuth, error) {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET環境変数が設定されていません")
	}

	return &JWTAuth{
		jwtSecret: []byte(secret),
		repo:      repo,
	}, nil
}

func (j *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "無効な認証ヘッダー形式です", http.StatusUnauthorized)
			return
		}

		tokenString := tokenParts[1]

		token, err := jwt.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("予期しない署名方式: %v", token.Header["alg"])
			}
			return j.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "無効なトークンです", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*SupabaseJWTClaims)
		if !ok {
			http.Error(w, "トークンのクレームが無効です", http.StatusUnauthorized)
			return
		}

		user := &AuthUser{
			ID:       claims.Subject,
			Email:    claims.Email,
			Metadata: claims.UserMetadata,
		}

		// ユーザー情報をコンテキストに保存
		ctx := context.WithValue(r.Context(), UserContextKey, user)

		// DBからユーザー情報を取得してコンテキストに保存（同期的に実行）
		if j.repo != nil {
			if dbUser := j.loadDBUser(user); dbUser != nil {
				ctx = context.WithValue(ctx, DBUserContextKey, dbUser)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (j *JWTAuth) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
			tokenString := tokenParts[1]

			token, err := jwt.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("予期しない署名方式: %v", token.Header["alg"])
				}
				return j.jwtSecret, nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*SupabaseJWTClaims); ok {
					user := &AuthUser{
						ID:       claims.Subject,
						Email:    claims.Email,
						Metadata: claims.UserMetadata,
					}

					// ユーザー情報をコンテキストに保存
					ctx := context.WithValue(r.Context(), UserContextKey, user)

					// DBからユーザー情報を取得してコンテキストに保存（同期的に実行）
					if j.repo != nil {
						if dbUser := j.loadDBUser(user); dbUser != nil {
							ctx = context.WithValue(ctx, DBUserContextKey, dbUser)
						}
					}

					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserFromContext(ctx context.Context) (*AuthUser, bool) {
	user, ok := ctx.Value(UserContextKey).(*AuthUser)
	return user, ok
}

// GetDBUserFromContext はコンテキストからDB上のユーザー情報を取得
func GetDBUserFromContext(ctx context.Context) (*models.User, bool) {
	dbUser, ok := ctx.Value(DBUserContextKey).(*models.User)
	return dbUser, ok
}

// loadDBUser はDBからユーザー情報を同期的に取得（存在しない場合はnil）
func (j *JWTAuth) loadDBUser(authUser *AuthUser) *models.User {
	supabaseUserID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil
	}

	existingUser, err := j.repo.User.FindUserBySupabaseUserID(supabaseUserID)
	if err != nil || existingUser == nil {
		return nil
	}

	return existingUser
}

// EnsureUserExists は新規ユーザーを作成（SyncUserエンドポイントから呼ばれる）
func (j *JWTAuth) EnsureUserExists(authUser *AuthUser) (*models.User, error) {
	supabaseUserID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid Supabase user ID: %v", err)
	}

	// 既存ユーザーをチェック
	existingUser, err := j.repo.User.FindUserBySupabaseUserID(supabaseUserID)
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	// PSN IDを取得
	var psnOnlineID *string
	if authUser.Metadata != nil {
		if val, ok := authUser.Metadata["psn_id"].(string); ok && val != "" {
			psnOnlineID = &val
		}
	}

	// 表示名を生成
	displayName := authUser.Email
	if idx := strings.Index(authUser.Email, "@"); idx > 0 {
		displayName = authUser.Email[:idx]
	}

	// 新規ユーザーを作成
	now := time.Now()
	newUser := &models.User{
		SupabaseUserID: supabaseUserID,
		Email:          authUser.Email,
		DisplayName:    displayName,
		PSNOnlineID:    psnOnlineID,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := j.repo.User.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	return newUser, nil
}
