package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

type RoomDetailHandler struct {
	BaseHandler
}

func NewRoomDetailHandler(repo *repository.Repository) *RoomDetailHandler {
	return &RoomDetailHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
	}
}

type RoomDetailPageData struct {
	Room    *models.Room         `json:"room"`
	Members []*models.RoomMember `json:"members"`
}

func (h *RoomDetailHandler) RoomDetail(w http.ResponseWriter, r *http.Request) {
	// URLパラメータから部屋IDを取得
	vars := mux.Vars(r)
	roomIDStr := vars["id"]

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "無効な部屋IDです", http.StatusBadRequest)
		return
	}

	// 部屋情報を取得
	room, err := h.repo.Room.FindRoomByID(roomID)
	if err != nil {
		http.Error(w, "部屋が見つかりません", http.StatusNotFound)
		return
	}

	// 部屋が削除済みまたは非アクティブの場合
	if !room.IsActive {
		http.Error(w, "この部屋は利用できません", http.StatusNotFound)
		return
	}

	// ホスト情報を取得
	host, err := h.repo.User.FindUserByID(room.HostUserID)
	if err != nil {
		http.Error(w, "ホスト情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	room.Host = *host

	// ゲームバージョン情報を取得
	gameVersion, err := h.repo.GameVersion.FindGameVersionByID(room.GameVersionID)
	if err != nil {
		http.Error(w, "ゲームバージョン情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	room.GameVersion = *gameVersion

	// 部屋のメンバーを取得（仮実装）
	members := []*models.RoomMember{}
	// TODO: 実際のメンバー取得メソッドを実装

	// メンバー配列を作成（4人分の枠を確保）
	memberSlots := make([]*models.RoomMember, 4)

	// メンバーのユーザーIDを収集
	var userIDs []uuid.UUID
	for _, member := range members {
		if member.PlayerNumber >= 1 && member.PlayerNumber <= 4 {
			userIDs = append(userIDs, member.UserID)
		}
	}

	// ユーザー情報を一括取得
	users, err := h.repo.User.FindUsersByIDs(userIDs)
	if err != nil {
		// エラーがあってもメンバー表示は続行
		users = []models.User{}
	}

	// ユーザー情報のマップを作成（高速検索用）
	userMap := make(map[uuid.UUID]models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メンバー情報にユーザー情報を設定
	for _, member := range members {
		if member.PlayerNumber >= 1 && member.PlayerNumber <= 4 {
			if user, exists := userMap[member.UserID]; exists {
				member.User = user
				member.DisplayName = user.DisplayName
				if user.Username != nil && *user.Username != "" {
					member.DisplayName = *user.Username
				}
			}
			memberSlots[member.PlayerNumber-1] = member
		}
	}

	// テンプレート用のデータを準備
	data := TemplateData{
		Title:   room.Name + " - 部屋詳細",
		HasHero: false,
		User:    r.Context().Value("user"),
		PageData: RoomDetailPageData{
			Room:    room,
			Members: memberSlots,
		},
	}

	// カスタムレンダリング関数
	renderRoomDetailTemplate(w, "room_detail.tmpl", data)
}

// renderRoomDetailTemplate は部屋詳細専用のテンプレートレンダリング関数
func renderRoomDetailTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	tmpl, err := template.ParseFiles(
		filepath.Join("templates", "layouts", "room_detail.tmpl"),
		filepath.Join("templates", "pages", templateName),
	)
	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "room_detail", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
