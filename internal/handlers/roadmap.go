package handlers

import (
	"fmt"
	"net/http"

	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"
)

type RoadmapHandler struct {
	BaseHandler
	articlesPath string
	generator    *info.Generator
}

func NewRoadmapHandler(repo *repository.Repository, generator *info.Generator) *RoadmapHandler {
	return &RoadmapHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
		generator:    generator,
	}
}

// loadArticles はJSONファイルから記事を読み込む
func (h *RoadmapHandler) loadArticles() (info.ArticleList, error) {
	articles, err := loadArticlesWithFallback(h.articlesPath, h.generator)
	if err != nil {
		return nil, fmt.Errorf("記事データの読み込みに失敗しました: %w", err)
	}
	return articles, nil
}

// Index はロードマップ一覧を表示する
func (h *RoadmapHandler) Index(w http.ResponseWriter, r *http.Request) {
	articles, err := h.loadArticles()
	if err != nil {
		http.Error(w, "ロードマップの読み込みに失敗しました", http.StatusInternalServerError)
		return
	}

	// ロードマップのみフィルタリング
	roadmaps := articles.FilterByCategory(info.ArticleTypeRoadmap)

	// ステータス順、日付順でソート（予定 → 開発中 → 完了の順）
	// 簡易実装：日付降順のみ
	roadmaps = roadmaps.SortByDateDesc()

	data := TemplateData{
		Title:      "開発ロードマップ",
		HideHeader: true,
		StaticPage: true,
		PageData: map[string]interface{}{
			"Roadmaps": roadmaps,
		},
	}

	renderTemplate(w, "roadmap/index.tmpl", data)
}
