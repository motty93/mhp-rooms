package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"mhp-rooms/internal/models"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
)

func (h *Handler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "ログイン",
	}
	renderTemplate(w, "login.html", data)
}

func (h *Handler) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "新規登録",
	}
	renderTemplate(w, "register.html", data)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	PSNId      string `json:"psnId"`
	PlayerName string `json:"playerName"`
	AgreeTerms bool   `json:"agreeTerms"`
}

type AuthResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Token   string      `json:"token,omitempty"`
	User    interface{} `json:"user,omitempty"`
}

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

	resp, err := h.supabase.Auth.SignInWithEmailPassword(req.Email, req.Password)
	if err != nil {
		log.Printf("ログインエラー: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "メールアドレスまたはパスワードが間違っています",
		})
		return
	}

	expireTime := time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	if req.Remember {
		expireTime = time.Now().Add(30 * 24 * time.Hour)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    resp.AccessToken,
		Path:     "/",
		Expires:  expireTime,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "sb-refresh-token",
		Value:    resp.RefreshToken,
		Path:     "/",
		Expires:  expireTime,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	user, err := h.repo.FindUserByEmail(req.Email)
	if err != nil || user == nil {
		supabaseUserID, _ := uuid.Parse(resp.User.ID.String())
		user = &models.User{
			Email:          req.Email,
			SupabaseUserID: supabaseUserID,
			DisplayName:    req.Email,
			IsActive:       true,
			Role:           "user",
		}
		if err := h.repo.CreateUser(user); err != nil {
			log.Printf("ユーザー作成エラー: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		User: map[string]interface{}{
			"id":          user.ID,
			"email":       user.Email,
			"displayName": user.DisplayName,
			"username":    user.Username,
		},
	})
}

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

	resp, err := h.supabase.Auth.Signup(types.SignupRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Printf("新規登録エラー: %v", err)
		message := "アカウントの作成に失敗しました"
		if err.Error() == "User already registered" {
			message = "このメールアドレスは既に登録されています"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: message,
		})
		return
	}

	supabaseUserID, _ := uuid.Parse(resp.User.ID.String())
	user := models.User{
		Email:          req.Email,
		Username:       &req.PSNId,
		DisplayName:    req.PlayerName,
		SupabaseUserID: supabaseUserID,
		IsActive:       true,
		Role:           "user",
	}

	if err := h.repo.CreateUser(&user); err != nil {
		log.Printf("ユーザー保存エラー: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "ユーザー情報の保存に失敗しました",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "アカウントが作成されました。メールで認証を完了してください。",
	})
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("sb-access-token")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Success: true,
			Message: "ログアウトしました",
		})
		return
	}
	if err := h.supabase.Auth.Logout(); err != nil {
		log.Printf("ログアウトエラー: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "sb-refresh-token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "ログアウトしました",
	})
}
