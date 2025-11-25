package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"
)

type BlogHandler struct {
	BaseHandler
	articlesPath string
	feedPath     string
	atomPath     string
	generator    *info.Generator
}

func NewBlogHandler(repo *repository.Repository, generator *info.Generator) *BlogHandler {
	return &BlogHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/blog/articles.json",
		feedPath:     "static/generated/blog/feed.xml",
		atomPath:     "static/generated/blog/atom.xml",
		generator:    generator,
	}
}

// loadArticles はJSONファイルから記事を読み込む
func (h *BlogHandler) loadArticles() (info.ArticleList, error) {
	articles, err := loadArticlesWithFallback(h.articlesPath, h.generator)
	if err != nil {
		return nil, fmt.Errorf("ブログ記事データの読み込みに失敗しました: %w", err)
	}
	return articles, nil
}

// List はブログ一覧を表示する
func (h *BlogHandler) List(w http.ResponseWriter, r *http.Request) {
	articles, err := h.loadArticles()
	if err != nil {
		http.Error(w, "記事の読み込みに失敗しました", http.StatusInternalServerError)
		return
	}

	// カテゴリーフィルタリング
	category := r.URL.Query().Get("category")
	var filteredArticles info.ArticleList

	switch category {
	case "community":
		filteredArticles = articles.FilterByCategory(info.ArticleTypeBlogCommunity)
	case "technical":
		filteredArticles = articles.FilterByCategory(info.ArticleTypeBlogTechnical)
	case "troubleshooting":
		filteredArticles = articles.FilterByCategory(info.ArticleTypeBlogTroubleshooting)
	default:
		// すべてのブログ記事
		filteredArticles = append(filteredArticles, articles.FilterByCategory(info.ArticleTypeBlogCommunity)...)
		filteredArticles = append(filteredArticles, articles.FilterByCategory(info.ArticleTypeBlogTechnical)...)
		filteredArticles = append(filteredArticles, articles.FilterByCategory(info.ArticleTypeBlogTroubleshooting)...)
		category = "all"
	}

	// 日付降順でソート
	filteredArticles = filteredArticles.SortByDateDesc()

	data := TemplateData{
		Title:      "HuntersHub通信",
		StaticPage: true,
		PageData: map[string]interface{}{
			"Articles":        filteredArticles,
			"CurrentCategory": category,
		},
	}

	renderTemplate(w, r, "blog/list.tmpl", data)
}

// Detail は個別のブログ記事を表示する
func (h *BlogHandler) Detail(w http.ResponseWriter, r *http.Request) {
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
		StaticPage: true,
		HideHeader: true,
		PageData: map[string]interface{}{
			"Article": foundArticle,
		},
	}

	renderTemplate(w, r, "blog/detail.tmpl", data)
}

// Feed はRSSフィードを返す
func (h *BlogHandler) Feed(w http.ResponseWriter, r *http.Request) {
	feedData, err := readGeneratedFile(h.feedPath, h.generator)
	if err != nil {
		http.Error(w, "フィードが見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write(feedData)
}

// AtomFeed はAtomフィードを返す
func (h *BlogHandler) AtomFeed(w http.ResponseWriter, r *http.Request) {
	feedData, err := readGeneratedFile(h.atomPath, h.generator)
	if err != nil {
		http.Error(w, "フィードが見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	w.Write(feedData)
}
