package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	
	"github.com/google/uuid"
	"mhp-rooms/internal/models"
)

func (h *AuthHandler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	// Supabase OAuth URL を環境変数から取得
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		http.Error(w, "Supabase URL not configured", http.StatusInternalServerError)
		return
	}
	
	// コールバックURLを設定
	redirectTo := "http://localhost:8080/auth/google/callback"
	if r.Host != "localhost:8080" {
		redirectTo = "https://" + r.Host + "/auth/google/callback"
	}
	
	// OAuth URLを構築
	params := url.Values{}
	params.Set("provider", "google")
	params.Set("redirect_to", redirectTo)
	
	authURL := fmt.Sprintf("%s/auth/v1/authorize?%s", supabaseURL, params.Encode())
	
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>認証処理中...</title>
	<script>
		const hash = window.location.hash.substring(1);
		const params = new URLSearchParams(hash);
		const accessToken = params.get('access_token');
		const refreshToken = params.get('refresh_token');
		
		if (accessToken) {
			fetch('/auth/session', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					access_token: accessToken,
					refresh_token: refreshToken
				})
			}).then(response => {
				if (response.ok) {
					return response.json();
				}
				throw new Error('セッション確立に失敗しました');
			}).then(data => {
				if (data.needsProfile) {
					window.location.href = '/auth/complete-profile';
				} else {
					window.location.href = '/rooms';
				}
			}).catch(error => {
				alert('認証エラー: ' + error.message);
				window.location.href = '/auth/login';
			});
		} else {
			const error = params.get('error_description') || 'Unknown error';
			alert('認証エラー: ' + error);
			window.location.href = '/auth/login';
		}
	</script>
</head>
<body>
	<div style="display: flex; justify-content: center; align-items: center; height: 100vh;">
		<div style="text-align: center;">
			<h2>認証処理中...</h2>
			<p>しばらくお待ちください</p>
		</div>
	</div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Supabase v0.0.4 では Auth.WithToken が使えないため、直接APIを呼び出す
	// ここでは仮実装として、トークンからユーザー情報を取得する処理を簡略化
	// 実際の実装ではSupabaseのAPIを直接呼び出すか、SDKをアップグレードする必要があります
	
	// 仮のユーザー情報（実際にはトークンを検証してユーザー情報を取得する必要があります）
	userEmail := "user@example.com" // 実際にはトークンから取得
	
	user, err := h.repo.FindUserByEmail(userEmail)
	if err != nil || user == nil {
		// 新規ユーザーの作成
		user = &models.User{
			Email:          userEmail,
			SupabaseUserID: uuid.New(), // 実際にはSupabaseから取得
			DisplayName:    userEmail,
			IsActive:       true,
			Role:           "user",
		}
		
		if err := h.repo.CreateUser(user); err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}
	
	// セッションクッキーの設定
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    req.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 30,
	})
	
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-refresh-token",
		Value:    req.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 30,
	})
	
	needsProfile := user.Username == nil || *user.Username == ""
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"needsProfile": needsProfile,
		"user": map[string]interface{}{
			"id":          user.ID,
			"email":       user.Email,
			"displayName": user.DisplayName,
		},
	})
}