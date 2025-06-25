package handlers

import (
	"net/http"
)

// GoogleAuthHandler handles Google OAuth initiation
func (h *Handler) GoogleAuthHandler(w http.ResponseWriter, r *http.Request) {
	// 一旦コメントアウト（将来実装用）
	
	/* Google OAuth実装（コメントアウト中）
	// パラメータ取得
	remember := r.URL.Query().Get("remember")
	redirectURI := r.URL.Query().Get("redirect_uri")
	authType := r.URL.Query().Get("type") // "register" or ""
	
	// Google OAuth設定
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		http.Error(w, "Google OAuth not configured", http.StatusInternalServerError)
		return
	}
	
	// 状態管理用のランダム文字列生成
	state := generateRandomString(32)
	
	// セッションに状態保存
	session, _ := store.Get(r, "auth-session")
	session.Values["oauth_state"] = state
	session.Values["remember"] = remember
	session.Values["auth_type"] = authType
	session.Save(r, w)
	
	// Google OAuth URL構築
	baseURL := "https://accounts.google.com/o/oauth2/auth"
	params := url.Values{
		"client_id":     {googleClientID},
		"redirect_uri":  {redirectURI},
		"scope":         {"openid email profile"},
		"response_type": {"code"},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}
	
	authURL := baseURL + "?" + params.Encode()
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	*/
	
	// 現在は準備中メッセージを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": "Google認証は準備中です"}`))
}

// GoogleCallbackHandler handles Google OAuth callback
func (h *Handler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// 一旦コメントアウト（将来実装用）
	
	/* Google OAuth実装（コメントアウト中）
	// 認証コード取得
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}
	
	// セッションから状態確認
	session, _ := store.Get(r, "auth-session")
	savedState, ok := session.Values["oauth_state"].(string)
	if !ok || savedState != state {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	
	// アクセストークン取得
	token, err := exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}
	
	// ユーザー情報取得
	userInfo, err := getUserInfoFromGoogle(token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	
	// ユーザー登録/ログイン処理
	remember := session.Values["remember"].(string) == "true"
	authType := session.Values["auth_type"].(string)
	
	if authType == "register" {
		// 新規登録処理
		user, err := h.repo.CreateUserFromGoogle(userInfo)
		if err != nil {
			http.Error(w, "Registration failed", http.StatusInternalServerError)
			return
		}
		// ログイン状態設定
		setAuthSession(w, user, remember)
		
		// PSN IDが未設定の場合はプロフィール補完ページへ
		if user.PSNOnlineID == nil || *user.PSNOnlineID == "" {
			// セッションクリア
			session.Values["oauth_state"] = nil
			session.Values["remember"] = nil
			session.Values["auth_type"] = nil
			session.Save(r, w)
			
			http.Redirect(w, r, "/auth/complete-profile", http.StatusTemporaryRedirect)
			return
		}
	} else {
		// ログイン処理
		user, err := h.repo.FindUserByEmail(userInfo.Email)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		// ログイン状態設定
		setAuthSession(w, user, remember)
	}
	
	// セッションクリア
	session.Values["oauth_state"] = nil
	session.Values["remember"] = nil
	session.Values["auth_type"] = nil
	session.Save(r, w)
	
	// 成功時リダイレクト
	http.Redirect(w, r, "/rooms", http.StatusTemporaryRedirect)
	*/
	
	// 現在は準備中メッセージを返す
	http.Error(w, "Google認証コールバックは準備中です", http.StatusNotImplemented)
}

/* ヘルパー関数（コメントアウト中）

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func exchangeCodeForToken(code string) (*oauth2.Token, error) {
	// OAuth2設定
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	
	return config.Exchange(context.Background(), code)
}

func getUserInfoFromGoogle(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	
	return &userInfo, nil
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func setAuthSession(w http.ResponseWriter, user *User, remember bool) {
	// JWT生成またはセッション設定
	// remember = trueの場合は30日、falseの場合は24時間
	duration := 24 * time.Hour
	if remember {
		duration = 30 * 24 * time.Hour
	}
	
	// 実装詳細は認証システムに依存
}

*/