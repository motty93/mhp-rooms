package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// ユーザー情報をコンテキストに格納するためのキー
type contextKey string

const UserContextKey contextKey = "user"

// SupabaseJWTClaims はSupabaseのJWTクレーム構造
type SupabaseJWTClaims struct {
	jwt.RegisteredClaims
	Email      string                 `json:"email"`
	Phone      string                 `json:"phone"`
	AppMetadata map[string]interface{} `json:"app_metadata"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
}

// AuthUser は認証されたユーザー情報
type AuthUser struct {
	ID       string
	Email    string
	Metadata map[string]interface{}
}

// JWTAuth はJWT認証ミドルウェアを提供
type JWTAuth struct {
	jwtSecret []byte
}

// NewJWTAuth は新しいJWT認証ミドルウェアを作成
func NewJWTAuth() (*JWTAuth, error) {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET環境変数が設定されていません")
	}
	
	return &JWTAuth{
		jwtSecret: []byte(secret),
	}, nil
}

// Middleware はJWT認証を行うミドルウェア
func (j *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorizationヘッダーからトークンを取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}
		
		// Bearer トークンの形式をチェック
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "無効な認証ヘッダー形式です", http.StatusUnauthorized)
			return
		}
		
		tokenString := tokenParts[1]
		
		// JWTトークンをパース・検証
		token, err := jwt.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 署名方式の確認
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("予期しない署名方式: %v", token.Header["alg"])
			}
			return j.jwtSecret, nil
		})
		
		if err != nil || !token.Valid {
			http.Error(w, "無効なトークンです", http.StatusUnauthorized)
			return
		}
		
		// クレームからユーザー情報を抽出
		claims, ok := token.Claims.(*SupabaseJWTClaims)
		if !ok {
			http.Error(w, "トークンのクレームが無効です", http.StatusUnauthorized)
			return
		}
		
		// ユーザー情報を作成
		user := &AuthUser{
			ID:       claims.Subject,
			Email:    claims.Email,
			Metadata: claims.UserMetadata,
		}
		
		// コンテキストにユーザー情報を格納
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalMiddleware はオプショナルなJWT認証を行うミドルウェア
// トークンがある場合は検証し、ない場合もリクエストを通す
func (j *JWTAuth) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorizationヘッダーを確認
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// トークンがない場合はそのまま次へ
			next.ServeHTTP(w, r)
			return
		}
		
		// トークンがある場合は検証を試みる
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
			tokenString := tokenParts[1]
			
			// JWTトークンをパース・検証
			token, err := jwt.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("予期しない署名方式: %v", token.Header["alg"])
				}
				return j.jwtSecret, nil
			})
			
			// 検証に成功した場合のみユーザー情報を設定
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*SupabaseJWTClaims); ok {
					user := &AuthUser{
						ID:       claims.Subject,
						Email:    claims.Email,
						Metadata: claims.UserMetadata,
					}
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					r = r.WithContext(ctx)
				}
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext はコンテキストからユーザー情報を取得
func GetUserFromContext(ctx context.Context) (*AuthUser, bool) {
	user, ok := ctx.Value(UserContextKey).(*AuthUser)
	return user, ok
}