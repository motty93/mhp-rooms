package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	
	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go"
	"mhp-rooms/internal/models"
)

func (h *Handler) GoogleAuthSupabaseHandler(w http.ResponseWriter, r *http.Request) {
	redirectTo := "http://localhost:8080/auth/callback"
	if r.Host != "localhost:8080" {
		redirectTo = "https://" + r.Host + "/auth/callback"
	}
	
	resp, err := h.supabase.Auth.SignInWithOAuth(gotrue.SignInWithOAuthOpts{
		Provider:   "google",
		RedirectTo: redirectTo,
		Options: map[string]string{
			"access_type": "offline",
			"prompt":      "consent",
		},
	})
	
	if err != nil {
		log.Printf("Google OAuth URL生成エラー: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Google認証の初期化に失敗しました",
		})
		return
	}
	
	http.Redirect(w, r, resp.URL, http.StatusTemporaryRedirect)
}

func (h *Handler) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) SessionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	authClient := h.supabase.Auth.WithToken(req.AccessToken)
	userResp, err := authClient.GetUser()
	if err != nil {
		log.Printf("ユーザー情報取得エラー: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	
	user, err := h.repo.FindUserByEmail(userResp.User.Email)
	if err != nil || user == nil {
		supabaseUserID, _ := uuid.Parse(userResp.User.ID.String())
		user = &models.User{
			Email:          userResp.User.Email,
			SupabaseUserID: supabaseUserID,
			DisplayName:    userResp.User.Email,
			IsActive:       true,
			Role:           "user",
		}
		
		if err := h.repo.CreateUser(user); err != nil {
			log.Printf("ユーザー作成エラー: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}
	
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