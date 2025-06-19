package handlers

import (
	"log"
	"net/http"

	"mhp-rooms/internal/database"
	"mhp-rooms/internal/models"
)

// RoomsPageData は部屋一覧ページのデータ構造体
type RoomsPageData struct {
	Rooms        []models.Room        `json:"rooms"`
	GameVersions []models.GameVersion `json:"game_versions"`
	Filter       string               `json:"filter"`
	Total        int64                `json:"total"`
}

func RoomsHandler(w http.ResponseWriter, r *http.Request) {
	// ゲームバージョンフィルターを取得
	filter := r.URL.Query().Get("game_version")
	
	// ゲームバージョン一覧を取得
	var gameVersions []models.GameVersion
	result := database.DB.Where("is_active = ?", true).Order("display_order").Find(&gameVersions)
	if result.Error != nil {
		log.Printf("ゲームバージョン取得エラー: %v", result.Error)
		http.Error(w, "ゲームバージョンの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// ルーム一覧を取得（関連データも含む）
	var rooms []models.Room
	query := database.DB.Preload("GameVersion").Preload("Host").
		Where("is_active = ? AND status IN (?)", true, []string{"waiting", "playing"}).
		Order("created_at DESC")

	// フィルター適用
	if filter != "" {
		var gameVersion models.GameVersion
		if err := database.DB.Where("code = ? AND is_active = ?", filter, true).First(&gameVersion).Error; err == nil {
			query = query.Where("game_version_id = ?", gameVersion.ID)
		}
	}

	result = query.Find(&rooms)
	if result.Error != nil {
		log.Printf("ルーム取得エラー: %v", result.Error)
		http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 総件数を取得
	var total int64
	countQuery := database.DB.Model(&models.Room{}).Where("is_active = ? AND status IN (?)", true, []string{"waiting", "playing"})
	if filter != "" {
		var gameVersion models.GameVersion
		if err := database.DB.Where("code = ? AND is_active = ?", filter, true).First(&gameVersion).Error; err == nil {
			countQuery = countQuery.Where("game_version_id = ?", gameVersion.ID)
		}
	}
	countQuery.Count(&total)

	pageData := RoomsPageData{
		Rooms:        rooms,
		GameVersions: gameVersions,
		Filter:       filter,
		Total:        total,
	}

	data := TemplateData{
		Title:    "部屋一覧",
		HasHero:  false, // 部屋一覧ページはヒーローセクションがない
		PageData: pageData,
	}
	renderTemplate(w, "rooms.html", data)
}