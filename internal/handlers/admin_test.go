package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
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

func TestNormalizeAdminUserListParams(t *testing.T) {
	longQuery := strings.Repeat("あ", 101)
	tests := []struct {
		name      string
		query     string
		sort      string
		wantQuery string
		wantSort  string
	}{
		{"前後の空白を除去", "  hunter  ", "rooms", "hunter", "rooms"},
		{"100文字に制限", longQuery, "reports", strings.Repeat("あ", 100), "reports"},
		{"不正な並び順は登録日順", "", "unknown", "", "joined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAdminUserListParams(tt.query, tt.sort)
			if got.Query != tt.wantQuery || got.Sort != tt.wantSort {
				t.Fatalf("normalizeAdminUserListParams() = %+v, want query=%q sort=%q", got, tt.wantQuery, tt.wantSort)
			}
		})
	}
}

func TestAdminListPaginationQueries(t *testing.T) {
	userParams := repository.AdminUserListParams{Query: "room master", Sort: "joined"}
	userURL := paginationURL("/admin/users", adminUsersQuery(userParams), 2)
	if want := "/admin/users?page=2&q=room+master&sort=joined"; userURL != want {
		t.Errorf("ユーザー一覧のページURL = %q, want %q", userURL, want)
	}

	hostUserID := uuid.New()
	roomURL := paginationURL("/admin/rooms", adminRoomsQuery(&hostUserID), 3)
	if want := "/admin/rooms?host_user_id=" + hostUserID.String() + "&page=3"; roomURL != want {
		t.Errorf("部屋一覧のページURL = %q, want %q", roomURL, want)
	}
}

func TestParseAdminRoomsHostUserID(t *testing.T) {
	validID := uuid.New()
	tests := []struct {
		name    string
		value   string
		wantNil bool
		wantErr bool
	}{
		{"未指定", "  ", true, false},
		{"有効なUUID", " " + validID.String() + " ", false, false},
		{"不正なUUID", "not-a-uuid", true, true},
		{"nil UUID", uuid.Nil.String(), true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAdminRoomsHostUserID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, tt.wantErr)
			}
			if (got == nil) != tt.wantNil {
				t.Fatalf("ID = %v, wantNil %t", got, tt.wantNil)
			}
			if got != nil && *got != validID {
				t.Errorf("ID = %s, want %s", got, validID)
			}
		})
	}
}

func TestAdminRoomsRejectsInvalidHostUserID(t *testing.T) {
	handler := &AdminHandler{}
	request := httptest.NewRequest(http.MethodGet, "/admin/rooms?host_user_id=not-a-uuid", nil)
	response := httptest.NewRecorder()

	handler.Rooms(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestAdminUsersRendersSearchResultsAndPreservesListParams(t *testing.T) {
	chdirRepoRoot(t)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Room{}, &models.UserActivity{}, &models.UserReport{}); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	for i := 0; i < adminUsersPerPage+1; i++ {
		username := fmt.Sprintf("hunter-%02d", i)
		user := &models.User{
			BaseModel:      models.BaseModel{ID: uuid.New(), CreatedAt: now.Add(time.Duration(i) * time.Minute)},
			SupabaseUserID: uuid.New(),
			Email:          fmt.Sprintf("hunter-%02d@example.test", i),
			Username:       &username,
			DisplayName:    fmt.Sprintf("Hunter %02d", i),
			IsActive:       i != adminUsersPerPage,
			Role:           models.RoleUser,
		}
		if err := db.Create(user).Error; err != nil {
			t.Fatal(err)
		}
	}

	handler := NewAdminHandler(repository.NewRepository(adminHandlerTestDB{conn: db}))
	request := httptest.NewRequest(http.MethodGet, "/admin/users?q=+HUNTER+&sort=joined", nil)
	response := httptest.NewRecorder()

	handler.Users(response, request)

	body := response.Body.String()
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body:\n%s", response.Code, truncate(body, 1500))
	}
	for _, want := range []string{
		"hunter-20@example.test", // 無効ユーザーも表示する
		`value="HUNTER"`,
		`<option value="joined" selected>`,
		`href="/admin/users?page=2&amp;q=HUNTER&amp;sort=joined"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("描画結果に %q が含まれていない", want)
		}
	}
}

// TestRenderAdminTemplates 管理画面テンプレートがパース・描画できることを確認する
func TestRenderAdminTemplates(t *testing.T) {
	chdirRepoRoot(t)

	roomID := uuid.New()
	userID := uuid.New()
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
					{ID: roomID, HostUserID: userID, Name: "テスト部屋", GameVersion: "MHP2G", HostName: "ホスト太郎", Players: "1/4", StatusLabel: "募集中", StatusClass: "bg-green-100 text-green-800", ReportCount: 2, CreatedAt: "2026-08-20 12:00"},
				},
				Pagination: newPagination(1, 1, adminRoomsPerPage, "/admin/rooms"),
			},
			want: []string{"全部屋一覧", "テスト部屋", "2件", "募集中", "/users/" + userID.String()},
		},
		{
			template: "admin_users.tmpl",
			data: adminUsersData{
				Users: []adminUserRow{{
					ID: userID, DisplayName: "ハンター太郎", Username: "hunter-taro", Email: "hunter@example.test", Role: models.RoleUser,
					CreatedAt: "2026-08-20 12:00", RoomCount: 3, ReportCount: 2, LastActivityAt: "2026-08-21 12:00",
				}},
				Query: "hunter", Sort: "rooms",
				Pagination: newPaginationWithQuery(60, 2, adminUsersPerPage, "/admin/users", url.Values{"q": {"hunter"}, "sort": {"rooms"}}),
			},
			want: []string{"ユーザーを検索", "hunter@example.test", "/users/" + userID.String(), "/admin/rooms?host_user_id=" + userID.String(), "page=3"},
		},
		{
			template: "admin_users.tmpl",
			data:     adminUsersData{Sort: "joined", Pagination: newPagination(0, 1, adminUsersPerPage, "/admin/users")},
			want:     []string{"該当するユーザーが見つかりません", "条件をリセット"},
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

type adminHandlerTestDB struct {
	conn *gorm.DB
}

func (d adminHandlerTestDB) GetConn() *gorm.DB { return d.conn }
func (d adminHandlerTestDB) Close() error      { return nil }
func (d adminHandlerTestDB) GetType() string   { return "sqlite" }
