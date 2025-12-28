package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/info"
	"mhp-rooms/internal/repository"
)

type PageHandler struct {
	BaseHandler
	articlesPath string
	generator    *info.Generator
}

func NewPageHandler(repo *repository.Repository, generator *info.Generator) *PageHandler {
	return &PageHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		articlesPath: "static/generated/info/articles.json",
		generator:    generator,
	}
}

func (h *PageHandler) loadArticles() (info.ArticleList, error) {
	return loadArticlesWithFallback(h.articlesPath, h.generator)
}

func (h *PageHandler) Hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded block">
    <strong>Hello World!</strong> Go + HTMX + Tailwind CSS + Air ホットリロードで動作しています！
</div>
<script>
    document.getElementById('hello-result').classList.remove('hidden');
</script>`)
}

// Robots responds with an environment-aware robots.txt
func (h *PageHandler) Robots(w http.ResponseWriter, r *http.Request) {
	siteURL := strings.TrimRight(config.GetEnv("SITE_URL", "http://localhost:8080"), "/")
	env := strings.ToLower(config.GetEnv("ENV", "development"))
	var builder strings.Builder
	builder.WriteString("# robots.txt for HuntersHub\n\n")
	builder.WriteString("User-agent: *\n")
	if env == "production" {
		builder.WriteString("Allow: /\n\n")
		builder.WriteString("Disallow: /admin/\n")
		builder.WriteString("Disallow: /api/\n")
		builder.WriteString("Disallow: /settings/\n")
		builder.WriteString("Disallow: /profile/edit\n")
		builder.WriteString("Crawl-delay: 1\n")
	} else {
		builder.WriteString("Disallow: /\n")
		builder.WriteString("# Crawling is disabled outside production\n")
	}
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Sitemap: %s/sitemap.xml\n", siteURL))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(builder.String()))
}
