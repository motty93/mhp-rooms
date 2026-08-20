package middleware

import (
	"net/http"
)

// RequireAdmin 管理者ロールのユーザーのみ許可するミドルウェア。
// 認証ミドルウェア（JWTAuth.Middleware）の後段に置く前提で、コンテキストの DB ユーザーを参照する。
// 管理画面の存在を外部に知らせないため、権限がない場合は 403 ではなく 404 を返す
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetDBUserFromContext(r.Context())
		if !ok || user == nil || !user.IsAdmin() {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
