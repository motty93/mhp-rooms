package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"
)

type InfoHandler struct {
	BaseHandler
	articlesPath string
	feedPath     string
	atomPath     string
	generator    *info.Generator
}

func NewInfoHandler(repo *repository.Repository, generator *info.Generator) *InfoHandler {
	return &InfoHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
		feedPath:     "static/generated/info/feed.xml",
		atomPath:     "static/generated/info/atom.xml",
		generator:    generator,
	}
}

// loadArticles はJSONファイルから記事を読み込む
func (h *InfoHandler) loadArticles() (info.ArticleList, error) {
	articles, err := loadArticlesWithFallback(h.articlesPath, h.generator)
	if err != nil {
		return nil, fmt.Errorf("記事データの読み込みに失敗しました: %w", err)
	}
	return articles, nil
}

// List は更新情報一覧を表示する
func (h *InfoHandler) List(w http.ResponseWriter, r *http.Request) {
	articles, err := h.loadArticles()
	if err != nil {
		http.Error(w, "記事の読み込みに失敗しました", http.StatusInternalServerError)
		return
	}

	// カテゴリーフィルタリング
	category := r.URL.Query().Get("category")
	var filteredArticles info.ArticleList

	switch category {
	case "news":
		filteredArticles = articles.FilterByCategory(info.ArticleTypeNews)
	case "release":
		filteredArticles = articles.FilterByCategory(info.ArticleTypeRelease)
	case "maintenance":
		filteredArticles = articles.FilterByCategory(info.ArticleTypeMaintenance)
	default:
		// すべての更新情報（ロードマップは除外）
		filteredArticles = append(filteredArticles, articles.FilterByCategory(info.ArticleTypeRelease)...)
		filteredArticles = append(filteredArticles, articles.FilterByCategory(info.ArticleTypeNews)...)
		filteredArticles = append(filteredArticles, articles.FilterByCategory(info.ArticleTypeMaintenance)...)
		category = "all"
	}

	// 日付降順でソート
	filteredArticles = filteredArticles.SortByDateDesc()

	data := TemplateData{
		Title:      "更新情報",
		HideHeader: true,
		StaticPage: true,
		PageData: map[string]interface{}{
			"Articles":        filteredArticles,
			"CurrentCategory": category,
		},
	}

	renderTemplate(w, r, "info/list.tmpl", data)
}

// Detail は個別の更新情報を表示する
func (h *InfoHandler) Detail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "記事が見つかりません", http.StatusNotFound)
		return
	}

	articles, err := h.loadArticles()
	if err != nil {
		http.Error(w, "記事の読み込みに失敗しました", http.StatusInternalServerError)
		return
	}

	// スラッグで記事を検索
	var foundArticle *info.Article
	for _, article := range articles {
		if article.Slug == slug {
			foundArticle = article
			break
		}
	}

	if foundArticle == nil {
		http.Error(w, "記事が見つかりません", http.StatusNotFound)
		return
	}

	data := TemplateData{
		Title:      foundArticle.Title,
		HideHeader: true,
		StaticPage: true,
		PageData: map[string]interface{}{
			"Article": foundArticle,
		},
	}

	renderTemplate(w, r, "info/detail.tmpl", data)
}

// Feed はRSSフィードを返す
func (h *InfoHandler) Feed(w http.ResponseWriter, r *http.Request) {
	feedData, err := readGeneratedFile(h.feedPath, h.generator)
	if err != nil {
		http.Error(w, "フィードが見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write(feedData)
}

// AtomFeed はAtomフィードを返す
func (h *InfoHandler) AtomFeed(w http.ResponseWriter, r *http.Request) {
	feedData, err := readGeneratedFile(h.atomPath, h.generator)
	if err != nil {
		http.Error(w, "フィードが見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	w.Write(feedData)
}
