package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/view"
)

const (
	adminLogsPerPage     = 50
	adminRoomsPerPage    = 20
	adminUsersPerPage    = 20
	adminMessagesPerPage = 50

	// AdminActionView 管理者による部屋閲覧の監査ログ用アクション
	AdminActionView = "admin_view"
)

// adminJST 管理画面の日時表示用タイムゾーン
var adminJST = time.FixedZone("JST", 9*60*60)

// adminActionLabels RoomLog の action に対応する日本語ラベル
var adminActionLabels = map[string]string{
	"create":          "部屋作成",
	"join":            "参加",
	"leave":           "退出",
	"kick":            "キック",
	"update_settings": "設定変更",
	"dismiss":         "解散",
	"auto_dismiss":    "自動解散",
	AdminActionView:   "管理者閲覧",
}

type AdminHandler struct {
	repo *repository.Repository
}

// adminLogRow ダッシュボードのタイムライン1行分
type adminLogRow struct {
	CreatedAt   string
	ActionLabel string
	RoomID      uuid.UUID
	RoomName    string
	ActorName   string
	Detail      string
}

// adminRoomRow 部屋一覧の1行分
type adminRoomRow struct {
	ID          uuid.UUID
	HostUserID  uuid.UUID
	Name        string
	GameVersion string
	HostName    string
	Players     string
	StatusLabel string
	StatusClass string
	ReportCount int64
	CreatedAt   string
}

// adminUserRow ユーザー一覧の1行分（メールアドレスは管理画面専用）
type adminUserRow struct {
	ID             uuid.UUID
	DisplayName    string
	Username       string
	Email          string
	Role           string
	CreatedAt      string
	RoomCount      int64
	ReportCount    int64
	LastActivityAt string
}

// adminMessageRow 部屋詳細のチャットログ1行分
type adminMessageRow struct {
	CreatedAt string
	UserName  string
	Message   string
}

// adminMemberRow 部屋詳細のメンバー1行分
type adminMemberRow struct {
	UserID   uuid.UUID
	Name     string
	IsHost   bool
	JoinedAt string
}

// adminDashboardData ダッシュボードの PageData
type adminDashboardData struct {
	Logs       []adminLogRow
	Pagination Pagination
}

// adminRoomsData 部屋一覧の PageData
type adminRoomsData struct {
	Rooms      []adminRoomRow
	Pagination Pagination
}

// adminUsersData ユーザー一覧の PageData
type adminUsersData struct {
	Users      []adminUserRow
	Query      string
	Sort       string
	Pagination Pagination
}

// adminRoomDetailData 部屋詳細の PageData
type adminRoomDetailData struct {
	Room        *models.Room
	StatusLabel string
	StatusClass string
	CreatedAt   string
	DismissedAt string
	Members     []adminMemberRow
	Messages    []adminMessageRow
	OlderCursor string
}

func NewAdminHandler(repo *repository.Repository) *AdminHandler {
	return &AdminHandler{repo: repo}
}

// Dashboard 全部屋の操作ログを新しい順に表示するタイムライン
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	page := parsePageParam(r)

	logs, err := h.repo.RoomLog.ListRecentLogs(adminLogsPerPage, (page-1)*adminLogsPerPage)
	if err != nil {
		http.Error(w, "ログの取得に失敗しました", http.StatusInternalServerError)
		return
	}
	total, err := h.repo.RoomLog.CountLogs()
	if err != nil {
		http.Error(w, "ログ件数の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	data := adminDashboardData{
		Logs:       buildAdminLogRows(logs),
		Pagination: newPagination(total, page, adminLogsPerPage, "/admin"),
	}

	view.Template(w, "admin_dashboard.tmpl", view.Data{
		Title:    "管理ダッシュボード",
		PageData: data,
	})
}

// Rooms 解散済みも含む全部屋の一覧
func (h *AdminHandler) Rooms(w http.ResponseWriter, r *http.Request) {
	page := parsePageParam(r)
	hostUserID, err := parseAdminRoomsHostUserID(r.URL.Query().Get("host_user_id"))
	if err != nil {
		http.Error(w, "host_user_id は有効なUUIDを指定してください", http.StatusBadRequest)
		return
	}

	rooms, err := h.repo.Room.GetAllRoomsForAdmin(adminRoomsPerPage, (page-1)*adminRoomsPerPage, hostUserID)
	if err != nil {
		http.Error(w, "部屋一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	total, err := h.repo.Room.CountAllRooms(hostUserID)
	if err != nil {
		http.Error(w, "部屋件数の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	hostIDs := make([]uuid.UUID, 0, len(rooms))
	seen := make(map[uuid.UUID]struct{}, len(rooms))
	for _, room := range rooms {
		if _, ok := seen[room.HostUserID]; ok {
			continue
		}
		seen[room.HostUserID] = struct{}{}
		hostIDs = append(hostIDs, room.HostUserID)
	}

	reportCounts, err := h.repo.Report.CountReportsByReportedUserIDs(hostIDs)
	if err != nil {
		// 通報数は補助情報のため、取得に失敗しても一覧表示は継続する
		log.Printf("管理画面: 通報数の取得に失敗しました: %v", err)
		reportCounts = map[uuid.UUID]int64{}
	}

	data := adminRoomsData{
		Rooms:      buildAdminRoomRows(rooms, reportCounts),
		Pagination: newPaginationWithQuery(total, page, adminRoomsPerPage, "/admin/rooms", adminRoomsQuery(hostUserID)),
	}

	view.Template(w, "admin_rooms.tmpl", view.Data{
		Title:    "部屋一覧（管理）",
		PageData: data,
	})
}

// Users は無効化済みも含め、管理者だけにユーザーと集計情報を表示します。
func (h *AdminHandler) Users(w http.ResponseWriter, r *http.Request) {
	params := normalizeAdminUserListParams(r.URL.Query().Get("q"), r.URL.Query().Get("sort"))
	page := parsePageParam(r)
	params.Limit = adminUsersPerPage
	params.Offset = (page - 1) * adminUsersPerPage

	users, err := h.repo.User.ListAdminUsers(params)
	if err != nil {
		http.Error(w, "ユーザー一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	total, err := h.repo.User.CountAdminUsers(params.Query)
	if err != nil {
		http.Error(w, "ユーザー件数の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	data := adminUsersData{
		Users:      buildAdminUserRows(users),
		Query:      params.Query,
		Sort:       params.Sort,
		Pagination: newPaginationWithQuery(total, page, adminUsersPerPage, "/admin/users", adminUsersQuery(params)),
	}
	view.Template(w, "admin_users.tmpl", view.Data{
		Title:    "ユーザー一覧（管理）",
		PageData: data,
	})
}

// RoomDetail 部屋情報・メンバー・チャットログの閲覧専用ビュー。閲覧は監査ログに記録する
func (h *AdminHandler) RoomDetail(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	room, err := h.repo.Room.FindRoomByID(roomID)
	if err != nil || room == nil {
		http.NotFound(w, r)
		return
	}

	members, err := h.repo.Room.GetRoomMembers(roomID)
	if err != nil {
		members = []models.RoomMember{}
	}

	var beforeID *uuid.UUID
	if beforeStr := r.URL.Query().Get("before"); beforeStr != "" {
		if id, err := uuid.Parse(beforeStr); err == nil {
			beforeID = &id
		}
	}

	// 新しい順で取得し、チャットとして読めるよう古い順に並べ直す
	messages, err := h.repo.RoomMessage.GetMessages(roomID, adminMessagesPerPage, beforeID)
	if err != nil {
		http.Error(w, "メッセージの取得に失敗しました", http.StatusInternalServerError)
		return
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// さらに古いログがある場合のカーソル（表示中の最古メッセージID）
	var olderCursor string
	if len(messages) == adminMessagesPerPage {
		olderCursor = messages[0].ID.String()
	}

	// 監査ログ: 誰がいつどの部屋を閲覧したかを記録する
	if admin, ok := middleware.GetDBUserFromContext(r.Context()); ok && admin != nil {
		auditLog := &models.RoomLog{
			RoomID: roomID,
			UserID: &admin.ID,
			Action: AdminActionView,
			Details: models.JSONB{
				Data: map[string]interface{}{
					"admin_name": adminUserName(&admin.ID, admin),
				},
			},
		}
		if err := h.repo.RoomLog.CreateLog(auditLog); err != nil {
			// 監査ログの書き込み失敗は閲覧自体は妨げないが、必ず痕跡を残す
			log.Printf("管理画面: 監査ログの記録に失敗しました room_id=%s admin_id=%s: %v", roomID, admin.ID, err)
		}
	}

	statusLabel, statusClass := adminRoomStatus(room)
	data := adminRoomDetailData{
		Room:        room,
		StatusLabel: statusLabel,
		StatusClass: statusClass,
		CreatedAt:   formatAdminTime(room.CreatedAt),
		Members:     buildAdminMemberRows(members),
		Messages:    buildAdminMessageRows(messages),
		OlderCursor: olderCursor,
	}
	if room.DismissedAt != nil {
		data.DismissedAt = formatAdminTime(*room.DismissedAt)
	}

	view.Template(w, "admin_room_detail.tmpl", view.Data{
		Title:    fmt.Sprintf("%s（管理）", room.Name),
		PageData: data,
	})
}

// buildAdminLogRows RoomLog を表示用の行に変換する
func buildAdminLogRows(logs []models.RoomLog) []adminLogRow {
	rows := make([]adminLogRow, 0, len(logs))
	for _, l := range logs {
		label, ok := adminActionLabels[l.Action]
		if !ok {
			label = l.Action
		}

		roomName := l.Room.Name
		if roomName == "" {
			roomName = "（削除済みの部屋）"
		}

		rows = append(rows, adminLogRow{
			CreatedAt:   formatAdminTime(l.CreatedAt),
			ActionLabel: label,
			RoomID:      l.RoomID,
			RoomName:    roomName,
			ActorName:   adminUserName(l.UserID, l.User),
			Detail:      adminLogDetail(l.Details),
		})
	}
	return rows
}

// buildAdminRoomRows Room を表示用の行に変換する（通報数はホストユーザーに対するもの）
func buildAdminRoomRows(rooms []models.Room, reportCounts map[uuid.UUID]int64) []adminRoomRow {
	rows := make([]adminRoomRow, 0, len(rooms))
	for i := range rooms {
		room := &rooms[i]
		statusLabel, statusClass := adminRoomStatus(room)
		rows = append(rows, adminRoomRow{
			ID:          room.ID,
			HostUserID:  room.HostUserID,
			Name:        room.Name,
			GameVersion: room.GameVersion.Code,
			HostName:    adminUserName(&room.HostUserID, &room.Host),
			Players:     fmt.Sprintf("%d/%d", room.CurrentPlayers, room.MaxPlayers),
			StatusLabel: statusLabel,
			StatusClass: statusClass,
			ReportCount: reportCounts[room.HostUserID],
			CreatedAt:   formatAdminTime(room.CreatedAt),
		})
	}
	return rows
}

func buildAdminUserRows(users []repository.AdminUser) []adminUserRow {
	rows := make([]adminUserRow, 0, len(users))
	for _, user := range users {
		username := "-"
		if user.Username != nil && *user.Username != "" {
			username = *user.Username
		}
		lastActivityAt := "活動なし"
		if user.LastActivityAt != nil {
			lastActivityAt = formatAdminTime(*user.LastActivityAt)
		}
		rows = append(rows, adminUserRow{
			ID:             user.ID,
			DisplayName:    user.DisplayName,
			Username:       username,
			Email:          user.Email,
			Role:           user.Role,
			CreatedAt:      formatAdminTime(user.CreatedAt),
			RoomCount:      user.RoomCount,
			ReportCount:    user.ReportCount,
			LastActivityAt: lastActivityAt,
		})
	}
	return rows
}

func normalizeAdminUserListParams(query, sort string) repository.AdminUserListParams {
	normalizedQuery := strings.TrimSpace(query)
	if runes := []rune(normalizedQuery); len(runes) > 100 {
		normalizedQuery = string(runes[:100])
	}
	if sort != "rooms" && sort != "reports" && sort != "joined" {
		sort = "joined"
	}
	return repository.AdminUserListParams{Query: normalizedQuery, Sort: sort}
}

func adminUsersQuery(params repository.AdminUserListParams) url.Values {
	values := url.Values{}
	if params.Query != "" {
		values.Set("q", params.Query)
	}
	values.Set("sort", params.Sort)
	return values
}

func parseAdminRoomsHostUserID(value string) (*uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("parse host user ID: %w", err)
	}
	if parsed == uuid.Nil {
		return nil, fmt.Errorf("host user ID must not be nil UUID")
	}
	return &parsed, nil
}

func adminRoomsQuery(hostUserID *uuid.UUID) url.Values {
	values := url.Values{}
	if hostUserID != nil {
		values.Set("host_user_id", hostUserID.String())
	}
	return values
}

func buildAdminMemberRows(members []models.RoomMember) []adminMemberRow {
	rows := make([]adminMemberRow, 0, len(members))
	for i := range members {
		m := &members[i]
		rows = append(rows, adminMemberRow{
			UserID:   m.UserID,
			Name:     adminUserName(&m.UserID, &m.User),
			IsHost:   m.IsHost,
			JoinedAt: formatAdminTime(m.JoinedAt),
		})
	}
	return rows
}

func buildAdminMessageRows(messages []models.RoomMessage) []adminMessageRow {
	rows := make([]adminMessageRow, 0, len(messages))
	for i := range messages {
		m := &messages[i]
		rows = append(rows, adminMessageRow{
			CreatedAt: formatAdminTime(m.CreatedAt),
			UserName:  adminUserName(&m.UserID, &m.User),
			Message:   m.Message,
		})
	}
	return rows
}

// adminRoomStatus 部屋の状態ラベルとバッジ用クラスを返す
func adminRoomStatus(room *models.Room) (string, string) {
	switch {
	case !room.IsActive && room.DismissReason != nil && *room.DismissReason == models.DismissReasonInactive:
		return "自動解散", "bg-gray-100 text-gray-600"
	case !room.IsActive:
		return "解散", "bg-gray-100 text-gray-600"
	case room.IsClosed:
		return "closed", "bg-yellow-100 text-yellow-800"
	default:
		return "募集中", "bg-green-100 text-green-800"
	}
}

// adminUserName 表示名 → ユーザー名 → ID 先頭8桁の順でフォールバックする
func adminUserName(userID *uuid.UUID, user *models.User) string {
	if user != nil {
		if user.DisplayName != "" {
			return user.DisplayName
		}
		if user.Username != nil && *user.Username != "" {
			return *user.Username
		}
	}
	if userID == nil {
		return "システム"
	}
	return userID.String()[:8]
}

// adminLogDetail RoomLog.Details の代表的なキーを表示用文字列にする
func adminLogDetail(details models.JSONB) string {
	data, ok := details.Data.(map[string]interface{})
	if !ok {
		return ""
	}

	parts := make([]string, 0, 2)
	for _, key := range []string{"user_name", "room_name", "reason", "admin_name"} {
		if v, ok := data[key].(string); ok && v != "" {
			parts = append(parts, v)
		}
	}

	detail := ""
	for i, p := range parts {
		if i > 0 {
			detail += " / "
		}
		detail += p
	}
	return detail
}

// formatAdminTime 管理画面用に JST で日時を整形する
func formatAdminTime(t time.Time) string {
	return t.In(adminJST).Format("2006-01-02 15:04")
}
