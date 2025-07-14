package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

// ConfigHandler は設定関連のハンドラー
type ConfigHandler struct{}

// NewConfigHandler は新しいConfigHandlerを作成
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// GetSupabaseConfig はフロントエンド用のSupabase設定を返す
func (h *ConfigHandler) GetSupabaseConfig(w http.ResponseWriter, r *http.Request) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseAnonKey == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   true,
			"message": "Supabase設定が未設定です。SUPABASE_URLとSUPABASE_ANON_KEYを設定してください。",
			"config": map[string]string{
				"url":     "",
				"anonKey": "",
			},
		})
		return
	}

	config := map[string]interface{}{
		"error": false,
		"config": map[string]string{
			"url":     supabaseURL,
			"anonKey": supabaseAnonKey,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, "設定の取得に失敗しました", http.StatusInternalServerError)
		return
	}
}
