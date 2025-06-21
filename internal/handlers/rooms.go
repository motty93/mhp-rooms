package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"mhp-rooms/internal/models"
)

type RoomsPageData struct {
	Rooms        []models.Room        `json:"rooms"`
	GameVersions []models.GameVersion `json:"game_versions"`
	Filter       string               `json:"filter"`
	Total        int64                `json:"total"`
}

func (h *Handler) RoomsHandler(w http.ResponseWriter, r *http.Request) {
	// ゲームバージョンフィルターを取得
	filter := r.URL.Query().Get("game_version")

	// ゲームバージョン一覧を取得
	gameVersions, err := h.repo.GetActiveGameVersions()
	if err != nil {
		log.Printf("ゲームバージョン取得エラー: %v", err)
		http.Error(w, "ゲームバージョンの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// フィルター用のゲームバージョンIDを取得
	var gameVersionID *uuid.UUID
	if filter != "" {
		gameVersion, err := h.repo.FindGameVersionByCode(filter)
		if err == nil && gameVersion != nil {
			gameVersionID = &gameVersion.ID
		}
	}

	// ルーム一覧を取得
	rooms, err := h.repo.GetActiveRooms(gameVersionID, 100, 0)
	if err != nil {
		log.Printf("ルーム取得エラー: %v", err)
		http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 総件数を計算（簡易的に）
	total := int64(len(rooms))

	pageData := RoomsPageData{
		Rooms:        rooms,
		GameVersions: gameVersions,
		Filter:       filter,
		Total:        total,
	}

	data := TemplateData{
		Title:    "部屋一覧",
		HasHero:  false,
		PageData: pageData,
	}
	renderTemplate(w, "rooms.html", data)
}

type CreateRoomRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	GameVersionID   string `json:"game_version_id"`
	MaxPlayers      int    `json:"max_players"`
	Password        string `json:"password"`
	QuestType       string `json:"quest_type"`
	TargetMonster   string `json:"target_monster"`
	RankRequirement string `json:"rank_requirement"`
}

func (h *Handler) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	gameVersionID, err := uuid.Parse(req.GameVersionID)
	if err != nil {
		http.Error(w, "無効なゲームバージョンIDです", http.StatusBadRequest)
		return
	}

	// TODO: 認証からユーザーIDを取得
	hostUserID := uuid.New() // 仮のユーザーID

	room := &models.Room{
		Name:          req.Name,
		GameVersionID: gameVersionID,
		HostUserID:    hostUserID,
		MaxPlayers:    req.MaxPlayers,
		Status:        "waiting",
		IsActive:      true,
	}

	if req.Description != "" {
		room.Description = &req.Description
	}
	if req.QuestType != "" {
		room.QuestType = &req.QuestType
	}
	if req.TargetMonster != "" {
		room.TargetMonster = &req.TargetMonster
	}
	if req.RankRequirement != "" {
		room.RankRequirement = &req.RankRequirement
	}

	if err := room.SetPassword(req.Password); err != nil {
		log.Printf("パスワード設定エラー: %v", err)
		http.Error(w, "パスワードの設定に失敗しました", http.StatusInternalServerError)
		return
	}

	if err := h.repo.CreateRoom(room); err != nil {
		log.Printf("ルーム作成エラー: %v", err)
		http.Error(w, "ルームの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

type JoinRoomRequest struct {
	Password string `json:"password"`
}

func (h *Handler) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "無効なルームIDです", http.StatusBadRequest)
		return
	}

	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// TODO: 認証からユーザーIDを取得
	userID := uuid.New() // 仮のユーザーID

	if err := h.repo.JoinRoom(roomID, userID, req.Password); err != nil {
		log.Printf("ルーム参加エラー: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "ルームに参加しました"}`))
}

func (h *Handler) LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "無効なルームIDです", http.StatusBadRequest)
		return
	}

	// TODO: 認証からユーザーIDを取得
	userID := uuid.New() // 仮のユーザーID

	if err := h.repo.LeaveRoom(roomID, userID); err != nil {
		log.Printf("ルーム退室エラー: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "ルームから退室しました"}`))
}
