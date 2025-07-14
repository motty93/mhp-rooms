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

		ctx := context.WithValue(r.Context(), UserContextKey, user)
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

							ctx := context.WithValue(r.Context(), UserContextKey, user)
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

func GetDBUserFromContext(ctx context.Context) (*models.User, bool) {
	dbUser, ok := ctx.Value(DBUserContextKey).(*models.User)
	return dbUser, ok
}

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

func (j *JWTAuth) EnsureUserExistsWithContext(ctx context.Context, authUser *AuthUser) (*models.User, error) {
	if dbUser, exists := GetDBUserFromContext(ctx); exists && dbUser != nil {
		return dbUser, nil
	}

	supabaseUserID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid Supabase user ID: %v", err)
	}

	existingUser, err := j.repo.User.FindUserBySupabaseUserID(supabaseUserID)
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	var psnOnlineID *string
	if authUser.Metadata != nil {
		if val, ok := authUser.Metadata["psn_id"].(string); ok && val != "" {
			psnOnlineID = &val
		}
	}

	displayName := authUser.Email
	if idx := strings.Index(authUser.Email, "@"); idx > 0 {
		displayName = authUser.Email[:idx]
	}

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

func (j *JWTAuth) EnsureUserExists(authUser *AuthUser) (*models.User, error) {
	return j.EnsureUserExistsWithContext(context.Background(), authUser)
}
