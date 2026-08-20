package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/view"
)

func TestAdminRoomStatus(t *testing.T) {
	inactive := models.DismissReasonInactive

	tests := []struct {
		name string
		room models.Room
		want string
	}{
		{"募集中", models.Room{IsActive: true}, "募集中"},
		{"closed", models.Room{IsActive: true, IsClosed: true}, "closed"},
		{"ホストによる解散", models.Room{IsActive: false}, "解散"},
		{"自動解散", models.Room{IsActive: false, DismissReason: &inactive}, "自動解散"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, class := adminRoomStatus(&tt.room)
			if label != tt.want {
				t.Errorf("label = %q, want %q", label, tt.want)
			}
			if class == "" {
				t.Error("class が空")
			}
		})
	}
}

func TestBuildAdminLogRows(t *testing.T) {
	userID := uuid.New()
	userName := "hunter"
	logs := []models.RoomLog{
		{
			BaseModel: models.BaseModel{CreatedAt: time.Date(2026, 8, 20, 3, 0, 0, 0, time.UTC)},
			RoomID:    uuid.New(),
			UserID:    &userID,
			Action:    "kick",
			Room:      models.Room{Name: "テスト部屋"},
			User:      &models.User{Username: &userName},
			Details:   models.JSONB{Data: map[string]interface{}{"user_name": "対象者"}},
		},
		{
			BaseModel: models.BaseModel{CreatedAt: time.Date(2026, 8, 20, 3, 0, 0, 0, time.UTC)},
			RoomID:    uuid.New(),
			Action:    "auto_dismiss", // UserID なし = システムによる操作
			Room:      models.Room{Name: "放置部屋"},
		},
		{
			BaseModel: models.BaseModel{CreatedAt: time.Date(2026, 8, 20, 3, 0, 0, 0, time.UTC)},
			RoomID:    uuid.New(),
			Action:    "unknown_action",
			Room:      models.Room{Name: "部屋"},
		},
	}

	rows := buildAdminLogRows(logs)
	if len(rows) != 3 {
		t.Fatalf("rows = %d 件, want 3", len(rows))
	}

	if rows[0].ActionLabel != "キック" || rows[0].ActorName != "hunter" || rows[0].Detail != "対象者" {
		t.Errorf("キック行が想定と異なる: %+v", rows[0])
	}
	// UTC 03:00 は JST 12:00 で表示される
	if rows[0].CreatedAt != "2026-08-20 12:00" {
		t.Errorf("CreatedAt = %q, want JST 表示", rows[0].CreatedAt)
	}
	if rows[1].ActionLabel != "自動解散" || rows[1].ActorName != "システム" {
		t.Errorf("自動解散行が想定と異なる: %+v", rows[1])
	}
	// 未知のアクションは生の値をそのまま表示する
	if rows[2].ActionLabel != "unknown_action" {
		t.Errorf("未知アクションのラベル = %q", rows[2].ActionLabel)
	}
}

func TestBuildAdminRoomRows(t *testing.T) {
	hostA := uuid.New()
	hostB := uuid.New()
	rooms := []models.Room{
		{
			BaseModel:      models.BaseModel{ID: uuid.New()},
			Name:           "通報あり部屋",
			HostUserID:     hostA,
			CurrentPlayers: 2,
			MaxPlayers:     4,
			IsActive:       true,
			GameVersion:    models.GameVersion{Code: "MHP2G"},
			Host:           models.User{DisplayName: "ホストA"},
		},
		{
			BaseModel:   models.BaseModel{ID: uuid.New()},
			Name:        "通報なし部屋",
			HostUserID:  hostB,
			IsActive:    true,
			GameVersion: models.GameVersion{Code: "MHP3"},
			Host:        models.User{DisplayName: "ホストB"},
		},
	}
	counts := map[uuid.UUID]int64{hostA: 3}

	rows := buildAdminRoomRows(rooms, counts)
	if len(rows) != 2 {
		t.Fatalf("rows = %d 件, want 2", len(rows))
	}
	if rows[0].ReportCount != 3 || rows[1].ReportCount != 0 {
		t.Errorf("通報数が想定と異なる: %d, %d", rows[0].ReportCount, rows[1].ReportCount)
	}
	if rows[0].Players != "2/4" || rows[0].GameVersion != "MHP2G" || rows[0].HostName != "ホストA" {
		t.Errorf("行の内容が想定と異なる: %+v", rows[0])
	}
}

func TestAdminUserName(t *testing.T) {
	id := uuid.New()
	username := "hunter"

	if got := adminUserName(&id, &models.User{DisplayName: "表示名"}); got != "表示名" {
		t.Errorf("表示名優先のはずが %q", got)
	}
	if got := adminUserName(&id, &models.User{Username: &username}); got != "hunter" {
		t.Errorf("ユーザー名フォールバックのはずが %q", got)
	}
	if got := adminUserName(&id, nil); got != id.String()[:8] {
		t.Errorf("ID 先頭8桁のはずが %q", got)
	}
	if got := adminUserName(nil, nil); got != "システム" {
		t.Errorf("システムのはずが %q", got)
	}
}

// TestRenderAdminTemplates 管理画面テンプレートがパース・描画できることを確認する
func TestRenderAdminTemplates(t *testing.T) {
	chdirRepoRoot(t)

	roomID := uuid.New()
	room := &models.Room{
		BaseModel:      models.BaseModel{ID: roomID},
		Name:           "テスト部屋",
		Description:    stringPtrForTest("説明文"),
		CurrentPlayers: 1,
		MaxPlayers:     4,
		IsActive:       true,
		GameVersion:    models.GameVersion{Code: "MHP2G"},
		Host:           models.User{DisplayName: "ホスト太郎"},
	}

	tests := []struct {
		template string
		data     interface{}
		want     []string
	}{
		{
			template: "admin_dashboard.tmpl",
			data: adminDashboardData{
				Logs: []adminLogRow{
					{CreatedAt: "2026-08-20 12:00", ActionLabel: "キック", RoomID: roomID, RoomName: "テスト部屋", ActorName: "hunter", Detail: "対象者"},
				},
				Pagination: newPagination(120, 2, adminLogsPerPage, "/admin"),
			},
			want: []string{"活動タイムライン", "テスト部屋", "キック", "/admin/rooms/" + roomID.String(), "前へ", "次へ"},
		},
		{
			template: "admin_rooms.tmpl",
			data: adminRoomsData{
				Rooms: []adminRoomRow{
					{ID: roomID, Name: "テスト部屋", GameVersion: "MHP2G", HostName: "ホスト太郎", Players: "1/4", StatusLabel: "募集中", StatusClass: "bg-green-100 text-green-800", ReportCount: 2, CreatedAt: "2026-08-20 12:00"},
				},
				Pagination: newPagination(1, 1, adminRoomsPerPage, "/admin/rooms"),
			},
			want: []string{"全部屋一覧", "テスト部屋", "2件", "募集中"},
		},
		{
			template: "admin_room_detail.tmpl",
			data: adminRoomDetailData{
				Room:        room,
				StatusLabel: "募集中",
				StatusClass: "bg-green-100 text-green-800",
				CreatedAt:   "2026-08-20 12:00",
				Members: []adminMemberRow{
					{UserID: uuid.New(), Name: "ホスト太郎", IsHost: true, JoinedAt: "2026-08-20 12:00"},
				},
				Messages: []adminMessageRow{
					{CreatedAt: "2026-08-20 12:01", UserName: "参加者", Message: "よろしく！"},
				},
				OlderCursor: uuid.New().String(),
			},
			want: []string{"チャットログ", "よろしく！", "ホスト太郎", "さらに古いログ"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			w := httptest.NewRecorder()
			view.Template(w, tt.template, view.Data{Title: "管理", PageData: tt.data})

			body := w.Body.String()
			if w.Code != 200 || strings.Contains(body, "Template parsing error") || strings.Contains(body, "Template execution error") {
				t.Fatalf("status = %d, body:\n%s", w.Code, truncate(body, 1500))
			}
			for _, want := range tt.want {
				if !strings.Contains(body, want) {
					t.Errorf("描画結果に %q が含まれていない", want)
				}
			}
		})
	}
}

func stringPtrForTest(s string) *string {
	return &s
}
