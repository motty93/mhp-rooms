package handlers

import (
	"net/http"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoomJoinHandler struct {
	BaseHandler
}

func NewRoomJoinHandler(repo *repository.Repository) *RoomJoinHandler {
	return &RoomJoinHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
	}
}

type RoomJoinPageData struct {
	Room        *RoomBasicInfo `json:"room"`
	IsJoined    bool           `json:"is_joined"`
	IsHost      bool           `json:"is_host"`
	HasPassword bool           `json:"has_password"`
}

type RoomBasicInfo struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	RoomCode    string             `json:"room_code"`
	GameVersion models.GameVersion `json:"game_version"`
	Host        models.User        `json:"host"`
	MaxPlayers  int                `json:"max_players"`
	HasPassword bool               `json:"has_password"`
}

// RoomJoinPage 部屋参加専用ページ（スケルトン）
// 認証チェックと最小限のクエリのみ実行
func (h *RoomJoinHandler) RoomJoinPage(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "無効な部屋IDです", http.StatusBadRequest)
		return
	}

	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		redirectURL := "/auth/login?redirect=" + r.URL.Path
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// 軽量クエリ：基本情報のみ取得
	room, err := h.repo.Room.FindRoomByID(roomID)
	if err != nil {
		http.Error(w, "部屋が見つかりません", http.StatusNotFound)
		return
	}

	if !room.IsActive {
		http.Error(w, "この部屋は利用できません", http.StatusNotFound)
		return
	}

	isJoined := h.repo.Room.IsUserJoinedRoom(roomID, dbUser.ID)
	if isJoined {
		http.Redirect(w, r, "/rooms/"+roomID.String(), http.StatusFound)
		return
	}

	isHost := dbUser.ID == room.HostUserID
	if isHost {
		http.Redirect(w, r, "/rooms/"+roomID.String(), http.StatusFound)
		return
	}

	basicInfo := &RoomBasicInfo{
		ID:          room.ID,
		Name:        room.Name,
		RoomCode:    room.RoomCode,
		GameVersion: room.GameVersion,
		Host:        room.Host,
		MaxPlayers:  room.MaxPlayers,
		HasPassword: room.HasPassword(),
	}

	data := TemplateData{
		Title:   room.Name + " - 部屋参加",
		HasHero: false,
		User:    r.Context().Value("user"),
		PageData: RoomJoinPageData{
			Room:        basicInfo,
			IsJoined:    isJoined,
			IsHost:      isHost,
			HasPassword: room.HasPassword(),
		},
	}

	renderTemplate(w, "room_join.tmpl", data)
}
