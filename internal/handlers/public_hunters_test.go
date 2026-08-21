package handlers

import (
	"net/http/httptest"
	"testing"
	"time"

	"mhp-rooms/internal/repository"

	"github.com/google/uuid"
)

func TestNormalizeHunterListParams(t *testing.T) {
	request := httptest.NewRequest("GET", "/users?q=%20%E3%83%86%E3%82%B9%E3%83%88%20&sort=invalid&page=0", nil)
	query, sort, page := normalizeHunterListParams(request)
	if query != "テスト" || sort != "recent" || page != 1 {
		t.Fatalf("正規化結果 = %q, %q, %d", query, sort, page)
	}
}

func TestFormatHunterActivityTime(t *testing.T) {
	if got := formatHunterActivityTime(time.Now().Add(-time.Minute).Format(time.RFC3339Nano)); got != "1分前" {
		t.Fatalf("相対時刻 = %q", got)
	}
	if got := formatHunterActivityTime("invalid"); got != "" {
		t.Fatalf("不正な日時 = %q", got)
	}
}

func TestRenderHunterListPartial(t *testing.T) {
	chdirRepoRoot(t)
	w := httptest.NewRecorder()
	data := HunterListData{
		Hunters:    []HunterListItem{{PublicHunter: repository.PublicHunter{ID: uuid.New(), DisplayName: "テスト"}}},
		Sort:       "recent",
		Page:       1,
		TotalPages: 1,
		Total:      1,
	}
	if err := renderPartialTemplate(w, "hunter_list", data); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || w.Body.Len() == 0 {
		t.Fatalf("一覧描画に失敗: %d", w.Code)
	}
}

func TestDefaultSitemapIncludesPublicHunterList(t *testing.T) {
	entries := defaultSitemapEntries("https://huntershub.net", time.Now())
	for _, entry := range entries {
		if entry.Loc == "https://huntershub.net/users" {
			return
		}
	}
	t.Fatal("公開ハンター一覧がサイトマップに含まれていません")
}

func TestRenderUsersPage(t *testing.T) {
	chdirRepoRoot(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users", nil)
	renderTemplate(w, r, "users.tmpl", TemplateData{Title: "ハンターを探す", PageData: HunterListData{Sort: "recent", Page: 1, TotalPages: 1}})
	if w.Code != 200 {
		t.Fatalf("users.tmpl の描画に失敗: %d: %s", w.Code, w.Body.String())
	}
}
