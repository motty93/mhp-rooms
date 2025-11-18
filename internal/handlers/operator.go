package handlers

import (
	"fmt"
	"net/http"

	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"
)

type OperatorHandler struct {
	BaseHandler
	articlesPath string
	generator    *info.Generator
}

func NewOperatorHandler(repo *repository.Repository, generator *info.Generator) *OperatorHandler {
	return &OperatorHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
		generator:    generator,
	}
}

// loadArticles はJSONファイルから記事を読み込む
func (h *OperatorHandler) loadArticles() (info.ArticleList, error) {
	articles, err := loadArticlesWithFallback(h.articlesPath, h.generator)
	if err != nil {
		return nil, fmt.Errorf("記事データの読み込みに失敗しました: %w", err)
	}
	return articles, nil
}

// Index は運営者情報ページを表示する
func (h *OperatorHandler) Index(w http.ResponseWriter, r *http.Request) {
	articles, err := h.loadArticles()
	if err != nil {
		http.Error(w, "記事の読み込みに失敗しました", http.StatusInternalServerError)
		return
	}

	// operatorカテゴリーの記事を取得
	operatorArticles := articles.FilterByCategory(info.ArticleTypeOperator)

	// 記事が存在しない場合
	if len(operatorArticles) == 0 {
		http.Error(w, "運営者情報が見つかりません", http.StatusNotFound)
		return
	}

	// 最初の1件を取得（単一ページ想定）
	article := operatorArticles[0]

	data := TemplateData{
		Title:      "運営者情報",
		HideHeader: true,
		StaticPage: true,
		PageData: map[string]interface{}{
			"Article": article,
		},
	}

	renderTemplate(w, r, "operator/index.tmpl", data)
}
