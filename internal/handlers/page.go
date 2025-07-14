package handlers

import (
	"fmt"
	"net/http"

	"mhp-rooms/internal/repository"
)

type PageHandler struct {
	BaseHandler
}

func NewPageHandler(repo *repository.Repository) *PageHandler {
	return &PageHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
	}
}

func (h *PageHandler) Terms(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "利用規約",
	}
	renderTemplate(w, "terms.tmpl", data)
}

func (h *PageHandler) Privacy(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "プライバシーポリシー",
	}
	renderTemplate(w, "privacy.tmpl", data)
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
