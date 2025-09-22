package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"mhp-rooms/internal/repository"
)

type GameVersionHandler struct {
	BaseHandler
}

func NewGameVersionHandler(repo *repository.Repository) *GameVersionHandler {
	return &GameVersionHandler{
		BaseHandler: BaseHandler{repo: repo},
	}
}

func (h *GameVersionHandler) GetActiveGameVersionsAPI(w http.ResponseWriter, r *http.Request) {
	versions, err := h.repo.GameVersion.GetActiveGameVersions()
	if err != nil {
		log.Printf("ゲームバージョン取得エラー: %v", err)
		http.Error(w, "ゲームバージョンの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

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
