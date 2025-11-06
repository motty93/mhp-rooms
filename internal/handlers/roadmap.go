package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"
)

type RoadmapHandler struct {
	BaseHandler
	articlesPath string
}

func NewRoadmapHandler(repo *repository.Repository) *RoadmapHandler {
	return &RoadmapHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
	}
}

// loadArticles はJSONファイルから記事を読み込む
func (h *RoadmapHandler) loadArticles() (info.ArticleList, error) {
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
		Title: "開発ロードマップ",
		PageData: map[string]interface{}{
			"Roadmaps": roadmaps,
		},
	}

	renderTemplate(w, "roadmap/index.tmpl", data)
}
