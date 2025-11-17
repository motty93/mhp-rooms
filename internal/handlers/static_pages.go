package handlers

import (
	"fmt"
	"net/http"

	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"
)

type StaticPageHandler struct {
	BaseHandler
	articlesPath string
	generator    *info.Generator
	category     info.ArticleType
	pageTitle    string
}

// NewGuideHandler は使い方ガイドハンドラーを作成する
func NewGuideHandler(repo *repository.Repository, generator *info.Generator) *StaticPageHandler {
	return &StaticPageHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
		generator:    generator,
		category:     info.ArticleTypeGuide,
		pageTitle:    "HuntersHubの使い方",
	}
}

// NewFAQHandler はFAQハンドラーを作成する
func NewFAQHandler(repo *repository.Repository, generator *info.Generator) *StaticPageHandler {
	return &StaticPageHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
		generator:    generator,
		category:     info.ArticleTypeFAQ,
		pageTitle:    "よくある質問",
	}
}

// NewTermsHandler は利用規約ハンドラーを作成する
func NewTermsHandler(repo *repository.Repository, generator *info.Generator) *StaticPageHandler {
	return &StaticPageHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
		generator:    generator,
		category:     info.ArticleTypeTerms,
		pageTitle:    "利用規約",
	}
}

// NewPrivacyHandler はプライバシーポリシーハンドラーを作成する
func NewPrivacyHandler(repo *repository.Repository, generator *info.Generator) *StaticPageHandler {
	return &StaticPageHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
		generator:    generator,
		category:     info.ArticleTypePrivacy,
		pageTitle:    "プライバシーポリシー",
	}
}

// loadArticles はJSONファイルから記事を読み込む
func (h *StaticPageHandler) loadArticles() (info.ArticleList, error) {
	articles, err := loadArticlesWithFallback(h.articlesPath, h.generator)
	if err != nil {
		return nil, fmt.Errorf("記事データの読み込みに失敗しました: %w", err)
	}
	return articles, nil
}

// Show は静的ページを表示する
func (h *StaticPageHandler) Show(w http.ResponseWriter, r *http.Request) {
	articles, err := h.loadArticles()
	if err != nil {
		http.Error(w, "ページの読み込みに失敗しました", http.StatusInternalServerError)
		return
	}

	// カテゴリーでフィルタリング
	filteredArticles := articles.FilterByCategory(h.category)

	// 最初の記事を取得（各カテゴリーには1つの記事のみ想定）
	if len(filteredArticles) == 0 {
		http.Error(w, "ページが見つかりません", http.StatusNotFound)
		return
	}

	page := filteredArticles[0]

	data := TemplateData{
		Title:      h.pageTitle,
		HideHeader: false, // 共通ヘッダーを表示
		StaticPage: true,  // 静的ページとして扱う（Alpine.js等を無効化）
		PageData: map[string]interface{}{
			"Page": page,
		},
	}

	renderTemplate(w, "static_page.tmpl", data)
}
