package handlers

import (
	"net/http"

	"mhp-rooms/internal/info"
)

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	// ブログ記事を取得（最新3件）
	var latestBlogArticles info.ArticleList
	articles, err := loadArticlesWithFallback("static/generated/blog/articles.json", nil)
	if err == nil {
		// 全てのブログカテゴリの記事を取得
		blogArticles := append(articles.FilterByCategory(info.ArticleTypeBlogCommunity),
			articles.FilterByCategory(info.ArticleTypeBlogTechnical)...)
		blogArticles = append(blogArticles,
			articles.FilterByCategory(info.ArticleTypeBlogTroubleshooting)...)

		// 公開済みの記事のみ、日付降順でソート
		blogArticles = blogArticles.ExcludeDrafts().SortByDateDesc()

		// 最新3件のみ取得
		if len(blogArticles) > 3 {
			latestBlogArticles = blogArticles[:3]
		} else {
			latestBlogArticles = blogArticles
		}
	}

	data := TemplateData{
		Title:      "ホーム",
		HasHero:    true,
		StaticPage: true,
		PageData: map[string]interface{}{
			"LatestBlogArticles": latestBlogArticles,
		},
	}
	renderTemplate(w, r, "home.tmpl", data)
}
