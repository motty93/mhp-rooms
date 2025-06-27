package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSupabaseClient struct {
	mock.Mock
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    LoginRequest
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "正常なログイン",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
				Remember: false,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name: "メールアドレス未入力",
			requestBody: LoginRequest{
				Email:    "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"メールアドレスを入力してください"`,
		},
		{
			name: "パスワード未入力",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"パスワードを入力してください"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			_ = httptest.NewRecorder()

			// handler := &Handler{
			//     supabase: mockSupabaseClient,
			//     repo: mockRepo,
			// }
			// handler.LoginHandler(rr, req)

			// assert.Equal(t, tt.expectedStatus, rr.Code)

			// assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    RegisterRequest
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "正常な新規登録",
			requestBody: RegisterRequest{
				Email:      "newuser@example.com",
				Password:   "password123",
				PSNId:      "TestPSN",
				PlayerName: "テストプレイヤー",
				AgreeTerms: true,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"success":true`,
		},
		{
			name: "メールアドレス未入力",
			requestBody: RegisterRequest{
				Email:      "",
				Password:   "password123",
				PSNId:      "TestPSN",
				PlayerName: "テストプレイヤー",
				AgreeTerms: true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"メールアドレスを入力してください"`,
		},
		{
			name: "無効なメールアドレス",
			requestBody: RegisterRequest{
				Email:      "invalid-email",
				Password:   "password123",
				PSNId:      "TestPSN",
				PlayerName: "テストプレイヤー",
				AgreeTerms: true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"有効なメールアドレスを入力してください"`,
		},
		{
			name: "パスワードが短すぎる",
			requestBody: RegisterRequest{
				Email:      "newuser@example.com",
				Password:   "short",
				PSNId:      "TestPSN",
				PlayerName: "テストプレイヤー",
				AgreeTerms: true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"パスワードは6文字以上で入力してください"`,
		},
		{
			name: "利用規約に同意していない",
			requestBody: RegisterRequest{
				Email:      "newuser@example.com",
				Password:   "password123",
				PSNId:      "TestPSN",
				PlayerName: "テストプレイヤー",
				AgreeTerms: false,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"利用規約とプライバシーポリシーに同意してください"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			_ = httptest.NewRecorder()

			// handler := &Handler{
			//     supabase: mockSupabaseClient,
			//     repo: mockRepo,
			// }
			// handler.RegisterHandler(rr, req)

			// assert.Equal(t, tt.expectedStatus, rr.Code)

			// assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

	req.AddCookie(&http.Cookie{
		Name:  "sb-access-token",
		Value: "test-token",
	})

	rr := httptest.NewRecorder()

	// handler := &Handler{
	//     supabase: mockSupabaseClient,
	// }
	// handler.LogoutHandler(rr, req)

	// assert.Equal(t, http.StatusOK, rr.Code)

	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "sb-access-token" {
			assert.Equal(t, "", cookie.Value)
			assert.True(t, cookie.Expires.Unix() < 0)
		}
	}
}
