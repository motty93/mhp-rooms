package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"mhp-rooms/internal/repository"
)

// GameVersionHandler はゲームバージョン関連のHTTPリクエストを処理
type GameVersionHandler struct {
	BaseHandler
}

// NewGameVersionHandler は新しいGameVersionHandlerインスタンスを作成
func NewGameVersionHandler(repo *repository.Repository) *GameVersionHandler {
	return &GameVersionHandler{
		BaseHandler: BaseHandler{repo: repo},
	}
}

// GetActiveGameVersionsAPI はアクティブなゲームバージョン一覧をJSONで返すAPIエンドポイント
func (h *GameVersionHandler) GetActiveGameVersionsAPI(w http.ResponseWriter, r *http.Request) {
	// アクティブなゲームバージョンを取得
	versions, err := h.repo.GameVersion.GetActiveGameVersions()
	if err != nil {
		log.Printf("ゲームバージョン取得エラー: %v", err)
		http.Error(w, "ゲームバージョンの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// JSONレスポンスを設定
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// レスポンス用の構造体
	response := struct {
		GameVersions interface{} `json:"game_versions"`
	}{
		GameVersions: versions,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("JSONエンコードエラー: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
}
