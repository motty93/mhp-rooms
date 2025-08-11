package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

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

	// 認証されたユーザーの場合、最適化されたメソッドを使用
	var enhancedRooms []interface{}
	dbUser, isAuthenticated := middleware.GetDBUserFromContext(r.Context())

	// デバッグログ
	if isAuthenticated {
		log.Printf("[DEBUG] Rooms: 認証済みユーザー: ID=%v, Email=%s", dbUser.ID, dbUser.Email)
	} else {
		log.Printf("[DEBUG] Rooms: 未認証ユーザー")
	}

	if isAuthenticated && dbUser != nil {
		// パフォーマンス最適化: 1つのクエリで参加状態を取得
		roomsWithJoinStatus, err := h.repo.Room.GetActiveRoomsWithJoinStatus(&dbUser.ID, nil, 100, 0)
		if err != nil {
			http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		for _, roomWithStatus := range roomsWithJoinStatus {
			roomData := map[string]interface{}{
				"id":               roomWithStatus.Room.ID,
				"room_code":        roomWithStatus.Room.RoomCode,
				"name":             roomWithStatus.Room.Name,
				"description":      roomWithStatus.Room.Description,
				"game_version_id":  roomWithStatus.Room.GameVersionID,
				"game_version":     roomWithStatus.Room.GameVersion,
				"host_user_id":     roomWithStatus.Room.HostUserID,
				"host":             roomWithStatus.Room.Host,
				"max_players":      roomWithStatus.Room.MaxPlayers,
				"current_players":  roomWithStatus.Room.CurrentPlayers,
				"target_monster":   roomWithStatus.Room.TargetMonster,
				"rank_requirement": roomWithStatus.Room.RankRequirement,
				"is_active":        roomWithStatus.Room.IsActive,
				"is_closed":        roomWithStatus.Room.IsClosed,
				"created_at":       roomWithStatus.Room.CreatedAt,
				"updated_at":       roomWithStatus.Room.UpdatedAt,
				"password_hash":    roomWithStatus.Room.PasswordHash,
				"has_password":     roomWithStatus.Room.HasPassword(),
				"is_joined":        roomWithStatus.IsJoined,
			}
			enhancedRooms = append(enhancedRooms, roomData)
		}
	} else {
		// 未認証ユーザーの場合は従来の方法
		rooms, err := h.repo.GetActiveRooms(nil, 100, 0)
		if err != nil {
			http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
			return
		}

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
				"has_password":     room.HasPassword(),
				"is_joined":        false,
			}
			enhancedRooms = append(enhancedRooms, roomData)
		}
	}

	total := int64(len(enhancedRooms))

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

	// ユーザーの部屋状態をチェック
	status, activeRoom, err := h.repo.Room.GetUserRoomStatus(hostUserID)
	if err != nil {
		log.Printf("部屋状態取得エラー: %v", err)
		http.Error(w, "部屋状態の確認に失敗しました", http.StatusInternalServerError)
		return
	}

	// ホスト中の場合は新しい部屋を作成できない
	if status == "HOST" {
		response := map[string]interface{}{
			"error":   "HOST_ROOM_ACTIVE",
			"message": "既にホストとして部屋を持っています",
			"room": map[string]interface{}{
				"id":        activeRoom.ID,
				"name":      activeRoom.Name,
				"room_code": activeRoom.RoomCode,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 参加中の場合は退出する（確認はフロントエンドで行う）
	if status == "GUEST" && activeRoom != nil {
		// 退出処理を実行
		if leaveErr := h.repo.Room.LeaveRoom(activeRoom.ID, hostUserID); leaveErr != nil {
			// 退出に失敗した場合は処理を中断
			http.Error(w, "現在の部屋からの退出に失敗しました", http.StatusInternalServerError)
			return
		}
	}

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

	// 作成成功時には部屋詳細URLを返す
	response := map[string]interface{}{
		"message": "ルームを作成しました",
		"room_id": room.ID.String(),
		"room":    room,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

type JoinRoomRequest struct {
	Password    string `json:"password"`
	ForceJoin   bool   `json:"forceJoin"`   // 強制参加フラグ（他の部屋から退出して参加）
	ConfirmJoin bool   `json:"confirmJoin"` // ブロック警告を確認済みで参加
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(chi.URLParam(r, "id"))
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

	// ユーザーの部屋状態をチェック
	status, currentRoom, err := h.repo.Room.GetUserRoomStatus(userID)
	if err != nil {
		log.Printf("部屋状態取得エラー: %v", err)
		http.Error(w, "部屋状態の確認に失敗しました", http.StatusInternalServerError)
		return
	}

	// ホスト中の場合は他の部屋に参加できない
	if status == "HOST" {
		response := map[string]interface{}{
			"error":   "HOST_CANNOT_JOIN",
			"message": "ホスト中は他の部屋に参加できません",
			"room": map[string]interface{}{
				"id":        currentRoom.ID,
				"name":      currentRoom.Name,
				"room_code": currentRoom.RoomCode,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	// ブロック関係のチェック
	room, err := h.repo.Room.FindRoomByID(roomID)
	if err != nil {
		http.Error(w, "ルームが見つかりません", http.StatusNotFound)
		return
	}

	// 1. ホストがユーザーをブロックしているかチェック
	isBlockedByHost, _, blockErr := h.repo.UserBlock.CheckBlockRelationship(userID, room.HostUserID)
	if blockErr != nil {
		log.Printf("ブロック関係の確認エラー: %v", blockErr)
		http.Error(w, "ブロック関係の確認に失敗しました", http.StatusInternalServerError)
		return
	}

	if isBlockedByHost {
		response := map[string]interface{}{
			"error":     "BLOCKED_BY_HOST",
			"message":   "このルームには参加できません",
			"blockType": "host_block",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 2. 既存メンバーとのブロック関係をチェック
	blockedMembers, blockErr := h.repo.UserBlock.CheckRoomMemberBlocks(userID, roomID)
	if blockErr != nil {
		log.Printf("ルームメンバーとのブロック関係確認エラー: %v", blockErr)
		http.Error(w, "ブロック関係の確認に失敗しました", http.StatusInternalServerError)
		return
	}

	if len(blockedMembers) > 0 {
		response := map[string]interface{}{
			"error":     "BLOCKED_BY_MEMBER",
			"message":   "ブロック関係により参加できません",
			"blockType": "member_block",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 3. ユーザーがホストをブロックしているかチェック（警告のみ）
	_, isBlockingHost, blockErr := h.repo.UserBlock.CheckBlockRelationship(userID, room.HostUserID)
	if blockErr != nil {
		log.Printf("ブロック関係の確認エラー: %v", blockErr)
		http.Error(w, "ブロック関係の確認に失敗しました", http.StatusInternalServerError)
		return
	}

	if isBlockingHost && !req.ConfirmJoin {
		response := map[string]interface{}{
			"warning":              "USER_BLOCKING_HOST",
			"message":              "ブロック関係があるユーザーが部屋にいます。参加しますか？",
			"blockType":            "user_block",
			"requiresConfirmation": true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// forceJoinフラグが設定されている場合は、先に現在の部屋から退出する
	if req.ForceJoin {
		// 現在参加している部屋があれば退出
		activeRoom, findErr := h.repo.Room.FindActiveRoomByUserID(userID)
		if findErr == nil && activeRoom != nil && activeRoom.ID != roomID {
			// 退出処理を実行
			if leaveErr := h.repo.Room.LeaveRoom(activeRoom.ID, userID); leaveErr != nil {
				// 退出に失敗した場合は処理を中断
				http.Error(w, "現在の部屋からの退出に失敗しました", http.StatusInternalServerError)
				return
			}
		}
	}

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
	roomID, err := uuid.Parse(chi.URLParam(r, "id"))
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
	roomID, err := uuid.Parse(chi.URLParam(r, "id"))
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

	// 認証されたユーザーの場合、最適化されたメソッドを使用
	var enhancedRooms []interface{}
	dbUser, isAuthenticated := middleware.GetDBUserFromContext(r.Context())

	if isAuthenticated && dbUser != nil {
		// パフォーマンス最適化: 1つのクエリで参加状態を取得
		roomsWithJoinStatus, err := h.repo.Room.GetActiveRoomsWithJoinStatus(&dbUser.ID, nil, 100, 0)
		if err != nil {
			http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		for _, roomWithStatus := range roomsWithJoinStatus {
			roomData := map[string]interface{}{
				"id":               roomWithStatus.Room.ID,
				"room_code":        roomWithStatus.Room.RoomCode,
				"name":             roomWithStatus.Room.Name,
				"description":      roomWithStatus.Room.Description,
				"game_version_id":  roomWithStatus.Room.GameVersionID,
				"host_user_id":     roomWithStatus.Room.HostUserID,
				"max_players":      roomWithStatus.Room.MaxPlayers,
				"current_players":  roomWithStatus.Room.CurrentPlayers,
				"target_monster":   roomWithStatus.Room.TargetMonster,
				"rank_requirement": roomWithStatus.Room.RankRequirement,
				"is_active":        roomWithStatus.Room.IsActive,
				"is_closed":        roomWithStatus.Room.IsClosed,
				"created_at":       roomWithStatus.Room.CreatedAt,
				"updated_at":       roomWithStatus.Room.UpdatedAt,
				"game_version":     roomWithStatus.Room.GameVersion,
				"host":             roomWithStatus.Room.Host,
				"has_password":     roomWithStatus.Room.HasPassword(),
				"is_joined":        roomWithStatus.IsJoined,
			}
			enhancedRooms = append(enhancedRooms, roomData)
		}
	} else {
		// 未認証ユーザーの場合は従来の方法
		rooms, err := h.repo.Room.GetActiveRooms(nil, 100, 0)
		if err != nil {
			http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
			return
		}

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
				"is_joined":        false,
			}
			enhancedRooms = append(enhancedRooms, roomData)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms":         enhancedRooms,
		"game_versions": gameVersions,
		"total":         len(enhancedRooms),
	})
}

// isUserJoinedRoom ユーザーが指定の部屋に参加しているかチェック
// GetCurrentRoom 現在参加中の部屋を取得するAPIエンドポイント
func (h *RoomHandler) GetCurrentRoom(w http.ResponseWriter, r *http.Request) {
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
		// 参加中の部屋がない場合
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"current_room": null}`))
		return
	}

	// 部屋情報をレスポンス用に整形
	roomData := map[string]interface{}{
		"id":               activeRoom.ID,
		"room_code":        activeRoom.RoomCode,
		"name":             activeRoom.Name,
		"description":      activeRoom.Description,
		"game_version_id":  activeRoom.GameVersionID,
		"game_version":     activeRoom.GameVersion,
		"host_user_id":     activeRoom.HostUserID,
		"host":             activeRoom.Host,
		"max_players":      activeRoom.MaxPlayers,
		"current_players":  activeRoom.CurrentPlayers,
		"target_monster":   activeRoom.TargetMonster,
		"rank_requirement": activeRoom.RankRequirement,
		"is_active":        activeRoom.IsActive,
		"is_closed":        activeRoom.IsClosed,
		"created_at":       activeRoom.CreatedAt,
		"updated_at":       activeRoom.UpdatedAt,
		"has_password":     activeRoom.HasPassword(),
	}

	response := map[string]interface{}{
		"current_room": roomData,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *RoomHandler) isUserJoinedRoom(roomID, userID uuid.UUID) bool {
	return h.repo.Room.IsUserJoinedRoom(roomID, userID)
}

// UpdateRoomRequest 部屋設定更新リクエスト
type UpdateRoomRequest = CreateRoomRequest

func (h *RoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "無効なルームIDです", http.StatusBadRequest)
		return
	}

	// 入力値の検証
	var req UpdateRoomRequest
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

	userID := dbUser.ID

	// 部屋の存在確認とホスト権限チェック
	room, err := h.repo.FindRoomByID(roomID)
	if err != nil {
		http.Error(w, "部屋が見つかりません", http.StatusNotFound)
		return
	}

	if room.HostUserID != userID {
		http.Error(w, "部屋のホストのみが設定を変更できます", http.StatusForbidden)
		return
	}

	// 部屋情報の更新
	room.Name = req.Name
	room.GameVersionID = gameVersionID
	room.MaxPlayers = req.MaxPlayers

	if req.Description != "" {
		room.Description = &req.Description
	} else {
		room.Description = nil
	}
	if req.TargetMonster != "" {
		room.TargetMonster = &req.TargetMonster
	} else {
		room.TargetMonster = nil
	}
	if req.RankRequirement != "" {
		room.RankRequirement = &req.RankRequirement
	} else {
		room.RankRequirement = nil
	}

	if err := room.SetPassword(req.Password); err != nil {
		http.Error(w, "パスワードの設定に失敗しました", http.StatusInternalServerError)
		return
	}

	if err := h.repo.UpdateRoom(room); err != nil {
		http.Error(w, "ルームの更新に失敗しました", http.StatusInternalServerError)
		return
	}

	// 更新成功のレスポンス
	response := map[string]interface{}{
		"message": "ルーム設定を更新しました",
		"room":    room,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *RoomHandler) DismissRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(chi.URLParam(r, "id"))
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

	// 部屋の存在確認とホスト権限チェック
	room, err := h.repo.FindRoomByID(roomID)
	if err != nil {
		http.Error(w, "部屋が見つかりません", http.StatusNotFound)
		return
	}

	if room.HostUserID != userID {
		http.Error(w, "部屋のホストのみが解散できます", http.StatusForbidden)
		return
	}

	// 部屋解散処理
	if err := h.repo.DismissRoom(roomID); err != nil {
		http.Error(w, "部屋の解散に失敗しました", http.StatusInternalServerError)
		return
	}

	// 解散メッセージをSSEで通知
	if h.hub != nil {
		dismissMessage := models.RoomMessage{
			ID:          uuid.New(),
			RoomID:      roomID,
			UserID:      userID,
			Message:     fmt.Sprintf("ルームがホスト（%s）によって解散されました", dbUser.DisplayName),
			MessageType: "system",
		}
		dismissMessage.User = *dbUser

		event := sse.Event{
			ID:   dismissMessage.ID.String(),
			Type: "room_dismissed",
			Data: dismissMessage,
		}
		h.hub.BroadcastToRoom(roomID, event)
	}

	// 解散成功のレスポンス
	response := map[string]interface{}{
		"message":  "ルームを解散しました",
		"redirect": "/rooms",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUserRoomStatus 現在のユーザーの部屋状態を取得
func (h *RoomHandler) GetUserRoomStatus(w http.ResponseWriter, r *http.Request) {
	// 認証情報からユーザーIDを取得
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていないか、ユーザー情報が見つかりません", http.StatusUnauthorized)
		return
	}

	userID := dbUser.ID

	// ユーザーの部屋状態を取得
	status, room, err := h.repo.Room.GetUserRoomStatus(userID)
	if err != nil {
		log.Printf("部屋状態取得エラー: %v", err)
		http.Error(w, "部屋状態の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// レスポンスを構築
	response := map[string]interface{}{
		"status": status,
		"room":   nil,
	}

	if room != nil {
		response["room"] = map[string]interface{}{
			"id":        room.ID,
			"name":      room.Name,
			"room_code": room.RoomCode,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
