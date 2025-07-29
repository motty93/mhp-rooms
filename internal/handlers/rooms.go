package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/sse"
)

type RoomHandler struct {
	BaseHandler
	hub *sse.Hub
}

func NewRoomHandler(repo *repository.Repository, hub *sse.Hub) *RoomHandler {
	return &RoomHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		hub: hub,
	}
}

type RoomsPageData struct {
	Rooms        []interface{}        `json:"rooms"`
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

	// 認証されたユーザーの場合、参加状態を含めるため拡張データを作成
	var enhancedRooms []interface{}
	dbUser, isAuthenticated := middleware.GetDBUserFromContext(r.Context())

	for _, room := range rooms {
		roomData := map[string]interface{}{
			"id":               room.ID,
			"room_code":        room.RoomCode,
			"name":             room.Name,
			"description":      room.Description,
			"game_version_id":  room.GameVersionID,
			"game_version":     room.GameVersion,
			"host_user_id":     room.HostUserID,
			"host":             room.Host,
			"max_players":      room.MaxPlayers,
			"current_players":  room.CurrentPlayers,
			"target_monster":   room.TargetMonster,
			"rank_requirement": room.RankRequirement,
			"is_active":        room.IsActive,
			"is_closed":        room.IsClosed,
			"created_at":       room.CreatedAt,
			"updated_at":       room.UpdatedAt,
			"password_hash":    room.PasswordHash,
		}

		// 認証済みユーザーの場合、参加状態を追加
		if isAuthenticated && dbUser != nil {
			roomData["is_joined"] = h.isUserJoinedRoom(room.ID, dbUser.ID)
		} else {
			roomData["is_joined"] = false
		}

		enhancedRooms = append(enhancedRooms, roomData)
	}

	total := int64(len(rooms))

	pageData := RoomsPageData{
		Rooms:        enhancedRooms,
		GameVersions: gameVersions,
		Filter:       filter,
		Total:        total,
	}

	data := TemplateData{
		Title:    "部屋一覧",
		HasHero:  false,
		PageData: pageData,
	}
	renderTemplate(w, "rooms.tmpl", data)
}

type CreateRoomRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	GameVersionID   string `json:"game_version_id"`
	MaxPlayers      int    `json:"max_players"`
	Password        string `json:"password"`
	TargetMonster   string `json:"target_monster"`
	RankRequirement string `json:"rank_requirement"`
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	// 入力値の検証
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// 必須フィールドの検証
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "ルーム名は必須です", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 {
		http.Error(w, "ルーム名は100文字以内で入力してください", http.StatusBadRequest)
		return
	}
	if req.MaxPlayers < 1 || req.MaxPlayers > 4 {
		http.Error(w, "最大プレイヤー数は1〜4人の間で設定してください", http.StatusBadRequest)
		return
	}

	gameVersionID, err := uuid.Parse(req.GameVersionID)
	if err != nil {
		http.Error(w, "無効なゲームバージョンIDです", http.StatusBadRequest)
		return
	}

	// 認証情報からユーザーIDを取得
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていないか、ユーザー情報が見つかりません", http.StatusUnauthorized)
		return
	}

	hostUserID := dbUser.ID

	room := &models.Room{
		Name:           req.Name,
		GameVersionID:  gameVersionID,
		HostUserID:     hostUserID,
		MaxPlayers:     req.MaxPlayers,
		IsActive:       true,
		CurrentPlayers: 1, // ホストを含めた初期人数
	}

	if req.Description != "" {
		room.Description = &req.Description
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

	// 認証情報からユーザーIDを取得
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていないか、ユーザー情報が見つかりません", http.StatusUnauthorized)
		return
	}

	userID := dbUser.ID

	if err := h.repo.Room.JoinRoom(roomID, userID, req.Password); err != nil {
		// 既に参加している場合は部屋詳細に遷移
		if strings.HasPrefix(err.Error(), "ALREADY_JOINED:") {
			response := map[string]interface{}{
				"message":  "既に参加しています。部屋に移動します。",
				"roomId":   roomID.String(),
				"redirect": fmt.Sprintf("/rooms/%s", roomID.String()),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
		// 他の部屋に既に参加している場合
		if strings.HasPrefix(err.Error(), "OTHER_ROOM_ACTIVE:") {
			http.Error(w, "既に別の部屋に参加しています", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 入室メッセージをSSEで通知
	if h.hub != nil {
		joinMessage := models.RoomMessage{
			ID:          uuid.New(),
			RoomID:      roomID,
			UserID:      userID,
			Message:     fmt.Sprintf("%sさんが入室しました", dbUser.DisplayName),
			MessageType: "system",
		}
		joinMessage.User = *dbUser

		event := sse.Event{
			ID:   joinMessage.ID.String(),
			Type: "system_message",
			Data: joinMessage,
		}
		h.hub.BroadcastToRoom(roomID, event)
	}

	// 参加成功時には部屋詳細URLを返す
	response := map[string]interface{}{
		"message":  "ルームに参加しました",
		"roomId":   roomID.String(),
		"redirect": fmt.Sprintf("/rooms/%s", roomID.String()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *RoomHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "無効なルームIDです", http.StatusBadRequest)
		return
	}

	// 認証情報からユーザーIDを取得
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていないか、ユーザー情報が見つかりません", http.StatusUnauthorized)
		return
	}

	userID := dbUser.ID

	if err := h.repo.LeaveRoom(roomID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "ルームから退室しました"}`))
}

func (h *RoomHandler) LeaveCurrentRoom(w http.ResponseWriter, r *http.Request) {
	// 認証情報からユーザーIDを取得
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていないか、ユーザー情報が見つかりません", http.StatusUnauthorized)
		return
	}

	userID := dbUser.ID

	// 現在参加しているアクティブな部屋を検索
	activeRoom, err := h.repo.Room.FindActiveRoomByUserID(userID)
	if err != nil {
		http.Error(w, "アクティブな部屋が見つかりません", http.StatusNotFound)
		return
	}

	// 部屋から退出
	if err := h.repo.Room.LeaveRoom(activeRoom.ID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "現在の部屋から退室しました"}`))
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

	// 認証情報からユーザーIDを取得
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていないか、ユーザー情報が見つかりません", http.StatusUnauthorized)
		return
	}

	currentUserID := dbUser.ID

	// ルームを取得してホストチェック
	room, err := h.repo.FindRoomByID(roomID)
	if err != nil {
		http.Error(w, "ルームが見つかりません", http.StatusNotFound)
		return
	}

	// ホスト権限チェック
	if currentUserID != room.HostUserID {
		http.Error(w, "ルームのホストのみが開閉状態を変更できます", http.StatusForbidden)
		return
	}

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

	rooms, err := h.repo.Room.GetActiveRooms(nil, 100, 0)
	if err != nil {
		http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 認証されたユーザーの場合、参加状態を含める
	var enhancedRooms []interface{}
	dbUser, isAuthenticated := middleware.GetDBUserFromContext(r.Context())

	for _, room := range rooms {
		roomData := map[string]interface{}{
			"id":               room.ID,
			"room_code":        room.RoomCode,
			"name":             room.Name,
			"description":      room.Description,
			"game_version_id":  room.GameVersionID,
			"host_user_id":     room.HostUserID,
			"max_players":      room.MaxPlayers,
			"current_players":  room.CurrentPlayers,
			"target_monster":   room.TargetMonster,
			"rank_requirement": room.RankRequirement,
			"is_active":        room.IsActive,
			"is_closed":        room.IsClosed,
			"created_at":       room.CreatedAt,
			"updated_at":       room.UpdatedAt,
			"game_version":     room.GameVersion,
			"host":             room.Host,
			"has_password":     room.HasPassword(),
		}

		// 認証済みユーザーの場合、参加状態を追加
		if isAuthenticated && dbUser != nil {
			// 簡単な参加チェック：RoomMemberを直接確認する代わりに、別の方法を使用
			roomData["is_joined"] = h.isUserJoinedRoom(room.ID, dbUser.ID)
		} else {
			roomData["is_joined"] = false
		}

		enhancedRooms = append(enhancedRooms, roomData)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms":         enhancedRooms,
		"game_versions": gameVersions,
		"total":         len(enhancedRooms),
	})
}

// isUserJoinedRoom ユーザーが指定の部屋に参加しているかチェック
func (h *RoomHandler) isUserJoinedRoom(roomID, userID uuid.UUID) bool {
	return h.repo.Room.IsUserJoinedRoom(roomID, userID)
}
