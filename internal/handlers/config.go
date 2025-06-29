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
	config := map[string]string{
		"url":     os.Getenv("SUPABASE_URL"),
		"anonKey": os.Getenv("SUPABASE_ANON_KEY"),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, "設定の取得に失敗しました", http.StatusInternalServerError)
		return
	}
}