package handlers

import (
	"fmt"
	"net/http"

	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"

	"github.com/go-chi/chi/v5"
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

	// ステータス順、日付順でソート（開発中 → 予定 → 完了の順）
	roadmaps = roadmaps.SortByStatusAndDate()

	data := TemplateData{
		Title:      "開発ロードマップ",
		HideHeader: true,
		StaticPage: true,
		PageData: map[string]interface{}{
			"Roadmaps": roadmaps,
		},
	}

	renderTemplate(w, r, "roadmap/index.tmpl", data)
}

// Detail は個別のロードマップ詳細を表示する
func (h *RoadmapHandler) Detail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "ロードマップが見つかりません", http.StatusNotFound)
		return
	}

	articles, err := h.loadArticles()
	if err != nil {
		http.Error(w, "ロードマップの読み込みに失敗しました", http.StatusInternalServerError)
		return
	}

	// ロードマップのみフィルタリングしてからスラッグで検索
	roadmaps := articles.FilterByCategory(info.ArticleTypeRoadmap)
	var foundArticle *info.Article
	for _, article := range roadmaps {
		if article.Slug == slug {
			foundArticle = article
			break
		}
	}

	if foundArticle == nil {
		http.Error(w, "ロードマップが見つかりません", http.StatusNotFound)
		return
	}

	data := TemplateData{
		Title:      foundArticle.Title + " - 開発ロードマップ",
		HideHeader: true,
		StaticPage: true,
		PageData: map[string]interface{}{
			"Article": foundArticle,
		},
	}

	renderTemplate(w, r, "roadmap/detail.tmpl", data)
}
