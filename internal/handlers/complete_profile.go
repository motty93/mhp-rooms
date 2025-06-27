package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// CompleteProfileRequest プロフィール補完リクエスト
type CompleteProfileRequest struct {
	PSNId string `json:"psnId" validate:"required,min=3,max=16"`
}

// CompleteProfilePageHandler プロフィール補完ページを表示
func (h *Handler) CompleteProfilePage(w http.ResponseWriter, r *http.Request) {
	// セッションからユーザー情報を取得
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	// ユーザーがすでにPSN IDを持っているか確認
	user, err := h.repo.FindUserByID(userID.(uuid.UUID))
	if err != nil {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// すでにPSN IDが設定されている場合はリダイレクト
	if user.PSNOnlineID != nil && *user.PSNOnlineID != "" {
		http.Redirect(w, r, "/rooms", http.StatusTemporaryRedirect)
		return
	}

	renderTemplate(w, "complete_profile.html", TemplateData{
		Title: "プロフィール設定",
		User:  user,
	})
}

// CompleteProfileHandler プロフィール補完処理
func (h *Handler) CompleteProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "認証が必要です",
		})
		return
	}

	var req CompleteProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "無効なリクエストです",
		})
		return
	}

	// バリデーション
	if len(req.PSNId) < 3 || len(req.PSNId) > 16 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "PSN IDは3〜16文字で入力してください",
		})
		return
	}

	// PSN IDの形式チェック（英数字、ハイフン、アンダースコアのみ）
	if !isValidPSNId(req.PSNId) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "PSN IDは英数字、ハイフン、アンダースコアのみ使用できます",
		})
		return
	}

	// ユーザー情報を取得
	user, err := h.repo.FindUserByID(userID.(uuid.UUID))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "ユーザー情報の取得に失敗しました",
		})
		return
	}

	// PSN IDを更新
	user.PSNOnlineID = &req.PSNId
	if err := h.repo.UpdateUser(user); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "プロフィールの更新に失敗しました",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "プロフィールが更新されました",
		"redirectUrl": "/rooms",
	})
}

// CurrentUserHandler 現在のユーザー情報を返すAPI
func (h *Handler) CurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "認証が必要です",
		})
		return
	}

	user, err := h.repo.FindUserByID(userID.(uuid.UUID))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "ユーザー情報の取得に失敗しました",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// isValidPSNId PSN IDの形式をチェック
func isValidPSNId(psnId string) bool {
	for _, char := range psnId {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}
	return true
}