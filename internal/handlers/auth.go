package handlers

import (
	"encoding/json"
	"net/http"
)

// LoginPageHandler renders the login page
func (h *Handler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "ログイン",
	}
	renderTemplate(w, "login.html", data)
}

// RegisterPageHandler renders the register page
func (h *Handler) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "新規登録",
	}
	renderTemplate(w, "register.html", data)
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

// RegisterRequest represents the register request payload
type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	PSNId      string `json:"psnId"`
	PlayerName string `json:"playerName"`
	AgreeTerms bool   `json:"agreeTerms"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Token   string      `json:"token,omitempty"`
	User    interface{} `json:"user,omitempty"`
}

// LoginHandler handles login API requests
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "無効なリクエストです",
		})
		return
	}

	// バリデーション
	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "メールアドレスを入力してください",
		})
		return
	}

	if req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "パスワードを入力してください",
		})
		return
	}

	// TODO: 実際の認証ロジックを実装
	// 現在は仮実装
	if req.Email == "test@example.com" && req.Password == "password" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Success: true,
			Token:   "dummy_token_" + req.Email,
			User: map[string]string{
				"email": req.Email,
				"name":  "テストユーザー",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: false,
		Message: "メールアドレスまたはパスワードが間違っています",
	})
}

// RegisterHandler handles register API requests
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "無効なリクエストです",
		})
		return
	}

	// バリデーション
	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "メールアドレスを入力してください",
		})
		return
	}

	if !h.isValidEmail(req.Email) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "有効なメールアドレスを入力してください",
		})
		return
	}

	if len(req.Password) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "パスワードは6文字以上で入力してください",
		})
		return
	}


	if req.PSNId == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "PSN IDを入力してください",
		})
		return
	}

	if req.PlayerName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "プレイヤーネームを入力してください",
		})
		return
	}

	if !req.AgreeTerms {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "利用規約とプライバシーポリシーに同意してください",
		})
		return
	}

	// TODO: 実際のユーザー登録ロジックを実装
	// 現在は仮実装
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "アカウントが作成されました",
	})
}