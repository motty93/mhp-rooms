package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"mhp-rooms/internal/models"

	"github.com/google/uuid"
)

// テンプレートは実行時にリポジトリルートからの相対パスで読み込まれるため、テストではルートへ移動する
func chdirRepoRoot(t *testing.T) {
	t.Helper()
	t.Chdir("../..")
}

func sampleUser() *models.User {
	return &models.User{
		BaseModel:   models.BaseModel{ID: uuid.New()},
		DisplayName: "テストユーザー",
	}
}

func sampleRooms(n int) []RoomSummary {
	rooms := make([]RoomSummary, 0, n)
	for i := 0; i < n; i++ {
		rooms = append(rooms, RoomSummary{
			ID:          uuid.New(),
			Name:        "テスト部屋",
			GameVersion: "MHP2G",
			PlayerCount: "1/4",
			Status:      "募集中",
			CreatedAt:   "2026/08/17",
			IsClickable: true,
		})
	}
	return rooms
}

// sampleAutoDismissedRoom 自動削除された部屋の表示データ
func sampleAutoDismissedRoom() RoomSummary {
	return RoomSummary{
		ID:          uuid.New(),
		Name:        "放置された部屋",
		GameVersion: "MHP3",
		PlayerCount: "0/4人",
		Status:      "自動削除",
		StatusColor: "text-orange-600",
		StatusNote:  "一定期間利用がなかったため、自動的に削除されました",
		CreatedAt:   "3日前",
	}
}

func TestRenderRoomsTabPartials(t *testing.T) {
	chdirRepoRoot(t)

	tests := []struct {
		name         string
		template     string
		pagination   Pagination
		wantPageLink string
		wantNoPaging bool
	}{
		{
			name:         "自分のプロフィール: 複数ページならページ送りを描画",
			template:     "profile_rooms",
			pagination:   newPagination(45, 2, tabPerPage, "/api/profile/rooms"),
			wantPageLink: `hx-get="/api/profile/rooms?page=3"`,
		},
		{
			name:         "他ユーザー: 複数ページならページ送りを描画",
			template:     "user_profile_rooms",
			pagination:   newPagination(45, 1, tabPerPage, "/api/users/abc/rooms"),
			wantPageLink: `hx-get="/api/users/abc/rooms?page=2"`,
		},
		{
			name:         "1ページに収まる場合はページ送りを描画しない",
			template:     "profile_rooms",
			pagination:   newPagination(5, 1, tabPerPage, "/api/profile/rooms"),
			wantNoPaging: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			data := roomsTabData{Rooms: sampleRooms(3), Pagination: tt.pagination}
			if err := renderPartialTemplate(w, tt.template, data); err != nil {
				t.Fatalf("renderPartialTemplate() error = %v", err)
			}

			body := w.Body.String()
			if tt.wantNoPaging {
				if strings.Contains(body, `aria-label="ページ送り"`) {
					t.Errorf("ページ送りが描画されないはずが描画されている:\n%s", body)
				}
				return
			}
			if !strings.Contains(body, tt.wantPageLink) {
				t.Errorf("ページ送りリンク %q が見つからない:\n%s", tt.wantPageLink, body)
			}
			if !strings.Contains(body, `hx-target="#tab-content"`) {
				t.Errorf("hx-target がタブコンテンツを指していない")
			}
		})
	}
}

func TestRenderActivityTabPartial(t *testing.T) {
	chdirRepoRoot(t)

	w := httptest.NewRecorder()
	data := activityTabData{
		Activities: []Activity{{Type: "room_create", Title: "部屋を作成", Icon: "fa-door-open", IconColor: "text-blue-500", TimeAgo: "1時間前"}},
		Pagination: newPagination(21, 1, tabPerPage, "/api/profile/activity"),
	}
	if err := renderPartialTemplate(w, "profile_activity", data); err != nil {
		t.Fatalf("renderPartialTemplate() error = %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, `hx-get="/api/profile/activity?page=2"`) {
		t.Errorf("次ページへのリンクが見つからない:\n%s", body)
	}
	if !strings.Contains(body, "部屋を作成") {
		t.Errorf("アクティビティ本文が描画されていない")
	}
}

func TestRenderProfilePagesWithRoomsTab(t *testing.T) {
	chdirRepoRoot(t)

	tests := []struct {
		name     string
		template string
		pageData interface{}
	}{
		{
			name:     "自分のプロフィールページ",
			template: "profile.tmpl",
			pageData: ProfileData{
				User:            sampleUser(),
				PlayTimes:       &models.PlayTimes{},
				IsOwnProfile:    true,
				Rooms:           sampleRooms(2),
				RoomsPagination: newPagination(30, 1, tabPerPage, "/api/profile/rooms"),
			},
		},
		{
			name:     "他ユーザーのプロフィールページ",
			template: "user_profile.tmpl",
			pageData: UserProfileData{
				User:            sampleUser(),
				PlayTimes:       &models.PlayTimes{},
				Rooms:           sampleRooms(2),
				RoomsPagination: newPagination(30, 1, tabPerPage, "/api/users/abc/rooms"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/profile", nil)
			renderTemplate(w, r, tt.template, TemplateData{Title: "test", PageData: tt.pageData})

			if w.Code != 200 {
				t.Fatalf("status = %d, body:\n%s", w.Code, w.Body.String())
			}
			body := w.Body.String()
			if strings.Contains(body, "Template parsing error") || strings.Contains(body, "Template execution error") {
				t.Fatalf("テンプレートエラー:\n%s", body)
			}
			if !strings.Contains(body, `id="profile-tabs"`) {
				t.Errorf("スクロール先の id=profile-tabs が無い")
			}
			if !strings.Contains(body, `?page=2"`) {
				t.Errorf("初期表示に2ページ目へのリンクが無い")
			}
		})
	}
}

func TestRenderRoomsTabShowsAutoDismissNote(t *testing.T) {
	chdirRepoRoot(t)

	for _, tmpl := range []string{"profile_rooms", "user_profile_rooms"} {
		t.Run(tmpl, func(t *testing.T) {
			w := httptest.NewRecorder()
			data := roomsTabData{
				Rooms:      append(sampleRooms(1), sampleAutoDismissedRoom()),
				Pagination: newPagination(2, 1, tabPerPage, "/api/profile/rooms"),
			}
			if err := renderPartialTemplate(w, tmpl, data); err != nil {
				t.Fatalf("renderPartialTemplate() error = %v", err)
			}

			body := w.Body.String()
			if !strings.Contains(body, "自動削除") {
				t.Errorf("「自動削除」ラベルが描画されていない:\n%s", body)
			}
			if strings.Count(body, "一定期間利用がなかったため") != 1 {
				t.Errorf("注記は自動削除の部屋にだけ 1 回描画されるはず:\n%s", body)
			}
		})
	}
}
