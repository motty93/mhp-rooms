package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"mhp-rooms/internal/models"
)

func TestRequireAdmin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireAdmin(next)

	tests := []struct {
		name string
		user *models.User
		want int
	}{
		{"コンテキストにユーザーなし（未認証）", nil, http.StatusNotFound},
		{"一般ユーザー", &models.User{Role: models.RoleUser}, http.StatusNotFound},
		{"ロールが空", &models.User{}, http.StatusNotFound},
		{"管理者", &models.User{Role: models.RoleAdmin}, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/admin", nil)
			if tt.user != nil {
				r = r.WithContext(context.WithValue(r.Context(), DBUserContextKey, tt.user))
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}
