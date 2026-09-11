package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mhp-rooms/internal/info"

	"github.com/go-chi/chi/v5"
)

func TestInfoAndRoadmapPagesRenderStaticHeader(t *testing.T) {
	chdirRepoRoot(t)

	articlesPath := writeStaticPageTestArticles(t)
	infoHandler := &InfoHandler{articlesPath: articlesPath}
	roadmapHandler := &RoadmapHandler{articlesPath: articlesPath}

	tests := []struct {
		name string
		path string
		slug string
		show http.HandlerFunc
	}{
		{name: "更新情報一覧", path: "/info", show: infoHandler.List},
		{name: "更新情報詳細", path: "/info/release-note", slug: "release-note", show: infoHandler.Detail},
		{name: "ロードマップ一覧", path: "/roadmap", show: roadmapHandler.Index},
		{name: "ロードマップ詳細", path: "/roadmap/roadmap-item", slug: "roadmap-item", show: roadmapHandler.Detail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.slug != "" {
				routeContext := chi.NewRouteContext()
				routeContext.URLParams.Add("slug", tt.slug)
				request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
			}

			response := httptest.NewRecorder()
			tt.show(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			body := response.Body.String()
			if !strings.Contains(body, `id="static-header"`) {
				t.Error("静的ページ用ヘッダーが表示されていません")
			}
			if !strings.Contains(body, "pt-16") {
				t.Error("固定ヘッダー分の上余白がありません")
			}
			if strings.Contains(body, "/static/js/auth-store.js") {
				t.Error("静的ページで認証スクリプトが読み込まれています")
			}
		})
	}
}

func writeStaticPageTestArticles(t *testing.T) string {
	t.Helper()

	articles := info.ArticleList{
		{
			Title:    "更新情報",
			Slug:     "release-note",
			Date:     time.Date(2026, time.September, 7, 8, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
			Category: info.ArticleTypeRelease,
			Summary:  "更新情報の概要",
			Content:  "<p>更新内容</p>",
		},
		{
			Title:    "ロードマップ",
			Slug:     "roadmap-item",
			Date:     time.Date(2026, time.September, 7, 8, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
			Category: info.ArticleTypeRoadmap,
			Summary:  "ロードマップの概要",
			Status:   "planned",
			Content:  "<p>予定内容</p>",
		},
	}

	data, err := json.Marshal(articles)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "articles.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
