package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/view"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	Room        *models.Room         `json:"room"`
	Members     []*models.RoomMember `json:"members"`
	Logs        []models.RoomLog     `json:"logs"`
	MemberCount int                  `json:"member_count"`
	IsHost      bool                 `json:"is_host"`
	OGImageURL  string               `json:"og_image_url"`
}

func (h *RoomDetailHandler) RoomDetail(w http.ResponseWriter, r *http.Request) {
	// URLパラメータから部屋IDを取得
	roomIDStr := chi.URLParam(r, "id")

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

	// プリロードで取得できなかった場合は異常と判断
	if room.Host.ID == uuid.Nil {
		http.Error(w, "ホスト情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if room.GameVersion.ID == uuid.Nil {
		http.Error(w, "ゲームバージョン情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 部屋のメンバーを取得
	members, err := h.repo.Room.GetRoomMembers(roomID)
	if err != nil {
		// エラーがあってもページ表示は続行
		members = []models.RoomMember{}
	}

	// メンバー配列を作成（4人分の枠を確保）
	memberSlots := make([]*models.RoomMember, 4)
	memberCount := 0

	// メンバー情報を設定
	for i := range members {
		if members[i].PlayerNumber >= 1 && members[i].PlayerNumber <= 4 {
			// DisplayNameを設定（display_name > username の優先順位）
			displayName := members[i].User.DisplayName
			// display_nameが空の場合はusernameを使用
			if displayName == "" && members[i].User.Username != nil && *members[i].User.Username != "" {
				displayName = *members[i].User.Username
			}
			members[i].DisplayName = displayName
			memberSlots[members[i].PlayerNumber-1] = &members[i]
			memberCount++
		}
	}

	// 部屋のログを取得
	logs, err := h.repo.Room.GetRoomLogs(roomID)
	if err != nil {
		// エラーがあってもページ表示は続行
		logs = []models.RoomLog{}
	}

	// 現在のユーザーがホストかどうかを判定
	isHost := false
	if dbUser, exists := middleware.GetDBUserFromContext(r.Context()); exists && dbUser != nil {
		isHost = dbUser.ID == room.HostUserID
	}

	ogImageURL := BuildOGPImageURL(room.ID, room.OGVersion)

	// テンプレート用のデータを準備
	data := TemplateData{
		Title:   room.Name + " - 部屋詳細",
		HasHero: false,
		User:    r.Context().Value("user"),
		SSEHost: config.AppConfig.Server.SSEHost, // SSEサーバーのホスト
		PageData: RoomDetailPageData{
			Room:        room,
			Members:     memberSlots,
			Logs:        logs,
			MemberCount: memberCount,
			IsHost:      isHost,
			OGImageURL:  ogImageURL,
		},
	}

	// カスタムレンダリング関数
	renderRoomDetailTemplate(w, r, "room_detail.tmpl", data)
}

// renderRoomDetailTemplate は部屋詳細専用のテンプレートレンダリング関数
func renderRoomDetailTemplate(w http.ResponseWriter, r *http.Request, templateName string, data TemplateData) {
	funcMap := view.TemplateFuncs()
	data = withCanonicalURL(r, data)

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "layouts", "room_detail.tmpl"),
		filepath.Join("templates", "pages", templateName),
		filepath.Join("templates", "components", "room_settings_modal.tmpl"),
		filepath.Join("templates", "components", "room_detail_script.tmpl"),
		filepath.Join("templates", "components", "share_modal.tmpl"),
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
