package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	supa "github.com/supabase-community/supabase-go"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

type RoomHandler struct {
	BaseHandler
}

func NewRoomHandler(repo *repository.Repository, supabaseClient *supa.Client) *RoomHandler {
	return &RoomHandler{
		BaseHandler: BaseHandler{
			repo:     repo,
			supabase: supabaseClient,
		},
	}
}

type RoomsPageData struct {
	Rooms        []models.Room        `json:"rooms"`
	GameVersions []models.GameVersion `json:"game_versions"`
	Filter       string               `json:"filter"`
	Total        int64                `json:"total"`
}

func (h *RoomHandler) Rooms(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("game_version")

	gameVersions, err := h.repo.GetActiveGameVersions()
	if err != nil {
		http.Error(w, "ゲームバージョンの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 常にすべてのアクティブな部屋を取得（フィルタリングはフロントエンドで行う）
	rooms, err := h.repo.GetActiveRooms(nil, 100, 0)
	if err != nil {
		http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 総件数を計算（フィルタリング前の全件数）
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

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "パスワードの設定に失敗しました", http.StatusInternalServerError)
		return
	}

	if err := h.repo.CreateRoom(room); err != nil {
		http.Error(w, "ルームの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

type JoinRoomRequest struct {
	Password string `json:"password"`
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "ルームに参加しました"}`))
}

func (h *RoomHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "無効なルームIDです", http.StatusBadRequest)
		return
	}

	// TODO: 認証からユーザーIDを取得
	userID := uuid.New() // 仮のユーザーID

	if err := h.repo.LeaveRoom(roomID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "ルームから退室しました"}`))
}

type ToggleRoomClosedRequest struct {
	IsClosed bool `json:"is_closed"`
}

func (h *RoomHandler) ToggleRoomClosed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "無効なルームIDです", http.StatusBadRequest)
		return
	}

	var req ToggleRoomClosedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// TODO: 認証からユーザーIDを取得してホストかどうかチェック
	// 現在は仮実装

	// ルームを取得してホストチェック
	_, err = h.repo.FindRoomByID(roomID)
	if err != nil {
		http.Error(w, "ルームが見つかりません", http.StatusNotFound)
		return
	}

	// TODO: 認証からのユーザーIDとroom.HostUserIDを比較
	// if currentUserID != room.HostUserID {
	//     http.Error(w, "ルームのホストのみが開閉状態を変更できます", http.StatusForbidden)
	//     return
	// }

	if err := h.repo.ToggleRoomClosed(roomID, req.IsClosed); err != nil {
		http.Error(w, "ルームの開閉状態変更に失敗しました", http.StatusInternalServerError)
		return
	}

	status := "開いた"
	if req.IsClosed {
		status = "閉じた"
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "ルームを%s状態にしました"}`, status)))
}

// GetAllRoomsAPIHandler APIエンドポイント：常に全データを返す
func (h *RoomHandler) GetAllRoomsAPI(w http.ResponseWriter, r *http.Request) {
	gameVersions, err := h.repo.GetActiveGameVersions()
	if err != nil {
		http.Error(w, "ゲームバージョンの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 常にすべてのアクティブな部屋を取得
	rooms, err := h.repo.GetActiveRooms(nil, 100, 0)
	if err != nil {
		http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms":        rooms,
		"game_versions": gameVersions,
		"total":        len(rooms),
	})
}
