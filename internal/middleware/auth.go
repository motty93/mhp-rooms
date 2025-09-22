package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"mhp-rooms/internal/config"
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

// UserCache は短時間のユーザーキャッシュ
type UserCache struct {
	users  map[uuid.UUID]*models.User
	mutex  sync.RWMutex
	expiry map[uuid.UUID]time.Time
}

func NewUserCache() *UserCache {
	return &UserCache{
		users:  make(map[uuid.UUID]*models.User),
		expiry: make(map[uuid.UUID]time.Time),
	}
}

func (c *UserCache) Get(userID uuid.UUID) (*models.User, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if expTime, exists := c.expiry[userID]; !exists || time.Now().After(expTime) {
		return nil, false
	}

	user, exists := c.users[userID]
	return user, exists
}

func (c *UserCache) Set(userID uuid.UUID, user *models.User, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.users[userID] = user
	c.expiry[userID] = time.Now().Add(ttl)
}

func (c *UserCache) Delete(userID uuid.UUID) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.users, userID)
	delete(c.expiry, userID)
}

func (c *UserCache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for userID, expTime := range c.expiry {
		if now.After(expTime) {
			delete(c.users, userID)
			delete(c.expiry, userID)
		}
	}
}

type JWTAuth struct {
	jwtSecret []byte
	repo      *repository.Repository
	userCache *UserCache
}

func (j *JWTAuth) GetUserCache() *UserCache {
	return j.userCache
}

func NewJWTAuth(repo *repository.Repository) (*JWTAuth, error) {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET環境変数が設定されていません")
	}

	userCache := NewUserCache()

	// 定期的にキャッシュをクリーンアップ
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				userCache.Cleanup()
			}
		}
	}()

	return &JWTAuth{
		jwtSecret: []byte(secret),
		repo:      repo,
		userCache: userCache,
	}, nil
}

func (j *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var tokenString string

		// まずAuthorizationヘッダーを確認
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				tokenString = tokenParts[1]
			} else {
				// htmxリクエストの場合はHX-Redirectヘッダーを使用
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/auth/login")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				// 通常のリクエストの場合はリダイレクト
				http.Redirect(w, r, "/auth/login", http.StatusFound)
				return
			}
		} else {
			// Authorizationヘッダーがない場合、クエリパラメータのtokenを確認（SSE用）
			tokenString = r.URL.Query().Get("token")
			if tokenString == "" {
				// クエリパラメータにもない場合、クッキーを確認（ブラウザからのAJAX用）
				if cookie, err := r.Cookie("sb-access-token"); err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				// htmxリクエストの場合はHX-Redirectヘッダーを使用
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/auth/login")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				// 通常のリクエストの場合はリダイレクト
				http.Redirect(w, r, "/auth/login", http.StatusFound)
				return
			}
		}
		if config.AppConfig.Debug.AuthLogs {
			log.Printf("AUTH DEBUG: %s %s - トークン長: %d文字", r.Method, r.URL.Path, len(tokenString))
		}

		token, err := jwt.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("予期しない署名方式: %v", token.Header["alg"])
			}
			return j.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			if config.AppConfig.Debug.AuthLogs {
				log.Printf("AUTH DEBUG: %s %s - トークン解析エラー: %v, Valid: %t", r.Method, r.URL.Path, err, token != nil && token.Valid)
			}
			// htmxリクエストの場合はHX-Redirectヘッダーを使用
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/auth/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// 通常のリクエストの場合はリダイレクト
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		claims, ok := token.Claims.(*SupabaseJWTClaims)
		if !ok {
			if config.AppConfig.Debug.AuthLogs {
				log.Printf("AUTH DEBUG: %s %s - クレーム変換エラー", r.Method, r.URL.Path)
			}
			// htmxリクエストの場合はHX-Redirectヘッダーを使用
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/auth/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// 通常のリクエストの場合はリダイレクト
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		if config.AppConfig.Debug.AuthLogs {
			log.Printf("AUTH DEBUG: %s %s - 認証成功 ユーザーID: %s", r.Method, r.URL.Path, claims.Subject)
		}

		user := &AuthUser{
			ID:       claims.Subject,
			Email:    claims.Email,
			Metadata: claims.UserMetadata,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, "request", r) // デバッグ用
		if j.repo != nil {
			if dbUser := j.loadDBUser(ctx, user); dbUser != nil {
				ctx = context.WithValue(ctx, DBUserContextKey, dbUser)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (j *JWTAuth) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 既にコンテキストにユーザー情報とDBユーザー情報の両方がある場合はスキップ
		if user, userExists := GetUserFromContext(r.Context()); userExists {
			if _, dbUserExists := GetDBUserFromContext(r.Context()); dbUserExists {
				next.ServeHTTP(w, r)
				return
			}
			// ユーザーはいるがDBユーザーがない場合は、DBユーザーのみ取得
			if j.repo != nil {
				ctx := context.WithValue(r.Context(), "request", r) // デバッグ用
				if dbUser := j.loadDBUser(ctx, user); dbUser != nil {
					ctx = context.WithValue(ctx, DBUserContextKey, dbUser)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
			return
		}

		// 複数の場所からトークンを取得を試行
		var tokenString string

		// 1. Authorizationヘッダーから取得
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				tokenString = tokenParts[1]
			}
		}

		// 2. クエリパラメータから取得（SSE用）
		if tokenString == "" {
			tokenString = r.URL.Query().Get("token")
		}

		// 3. クッキーから取得（SSR用）
		if tokenString == "" {
			if cookie, err := r.Cookie("sb-access-token"); err == nil {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			// デバッグログ: 認証情報が見つからない
			if config.AppConfig.Debug.AuthLogs {
				log.Printf("AUTH DEBUG: %s %s - 認証トークンが見つかりません（未認証として継続）", r.Method, r.URL.Path)
			}
			next.ServeHTTP(w, r)
			return
		}

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
				ctx = context.WithValue(ctx, "request", r) // デバッグ用
				if j.repo != nil {
					if dbUser := j.loadDBUser(ctx, user); dbUser != nil {
						ctx = context.WithValue(ctx, DBUserContextKey, dbUser)

						// デバッグログ: 認証成功
						if config.AppConfig.Debug.AuthLogs {
							log.Printf("AUTH DEBUG: %s %s - 認証成功 ユーザーID: %s, Email: %s",
								r.Method, r.URL.Path, dbUser.ID, dbUser.Email)
						}
					}
				}

				r = r.WithContext(ctx)
			}
		} else {
			// デバッグログ: トークン解析エラー
			if config.AppConfig.Debug.AuthLogs {
				log.Printf("AUTH DEBUG: %s %s - トークン解析エラー: %v", r.Method, r.URL.Path, err)
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

func (j *JWTAuth) loadDBUser(ctx context.Context, authUser *AuthUser) *models.User {
	// 既にコンテキストにDBユーザーが存在する場合は再利用
	if dbUser, exists := GetDBUserFromContext(ctx); exists && dbUser != nil {
		return dbUser
	}

	supabaseUserID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return nil
	}

	// キャッシュから取得を試行
	if cachedUser, exists := j.userCache.Get(supabaseUserID); exists {
		return cachedUser
	}

	existingUser, err := j.repo.User.FindUserBySupabaseUserID(supabaseUserID)
	if err != nil || existingUser == nil {
		// ユーザーが見つからない場合は自動作成を試行
		newUser, createErr := j.EnsureUserExistsWithContext(ctx, authUser)
		if createErr != nil {
			return nil
		}
		// キャッシュに保存（5分間）
		j.userCache.Set(supabaseUserID, newUser, 5*time.Minute)
		return newUser
	}

	// キャッシュに保存（5分間）
	j.userCache.Set(supabaseUserID, existingUser, 5*time.Minute)

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

	username := authUser.Email
	if idx := strings.Index(authUser.Email, "@"); idx > 0 {
		username = authUser.Email[:idx]
	}
	displayName := ""

	now := time.Now()
	newUser := &models.User{
		BaseModel: models.BaseModel{
			CreatedAt: now,
			UpdatedAt: now,
		},
		SupabaseUserID: supabaseUserID,
		Email:          authUser.Email,
		Username:       &username,
		DisplayName:    displayName,
		PSNOnlineID:    psnOnlineID,
		IsActive:       true,
	}

	if err := j.repo.User.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	return newUser, nil
}

func (j *JWTAuth) EnsureUserExists(authUser *AuthUser) (*models.User, error) {
	return j.EnsureUserExistsWithContext(context.Background(), authUser)
}
