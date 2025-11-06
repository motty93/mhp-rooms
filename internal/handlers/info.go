package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"
)

type InfoHandler struct {
	BaseHandler
	articlesPath string
}

func NewInfoHandler(repo *repository.Repository) *InfoHandler {
	return &InfoHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
	}
}

// loadArticles はJSONファイルから記事を読み込む
func (h *InfoHandler) loadArticles() (info.ArticleList, error) {
	data, err := os.ReadFile(h.articlesPath)
	if err != nil {
		return nil, fmt.Errorf("記事データの読み込みエラー: %w", err)
	}

	var articles info.ArticleList
	if err := json.Unmarshal(data, &articles); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w", err)
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
	case "release":
		filteredArticles = articles.FilterByCategory(info.ArticleTypeRelease)
	case "maintenance":
		filteredArticles = articles.FilterByCategory(info.ArticleTypeMaintenance)
	default:
		// すべての更新情報（ロードマップは除外）
		releaseArticles := articles.FilterByCategory(info.ArticleTypeRelease)
		maintenanceArticles := articles.FilterByCategory(info.ArticleTypeMaintenance)
		filteredArticles = append(releaseArticles, maintenanceArticles...)
		category = "all"
	}

	// 日付降順でソート
	filteredArticles = filteredArticles.SortByDateDesc()

	data := TemplateData{
		Title: "更新情報",
		PageData: map[string]interface{}{
			"Articles":        filteredArticles,
			"CurrentCategory": category,
		},
	}

	renderTemplate(w, "info/list.tmpl", data)
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
		Title: foundArticle.Title,
		PageData: map[string]interface{}{
			"Article": foundArticle,
		},
	}

	renderTemplate(w, "info/detail.tmpl", data)
}

// Feed はRSSフィードを返す
func (h *InfoHandler) Feed(w http.ResponseWriter, r *http.Request) {
	feedPath := "static/generated/info/feed.xml"
	feedData, err := os.ReadFile(feedPath)
	if err != nil {
		http.Error(w, "フィードが見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write(feedData)
}
