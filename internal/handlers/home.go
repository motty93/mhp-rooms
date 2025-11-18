package handlers

import "net/http"

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title:      "ホーム",
		HasHero:    true,
		StaticPage: true,
	}
	renderTemplate(w, r, "home.tmpl", data)
}
