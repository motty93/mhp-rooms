package handlers

import "net/http"

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title:   "ホーム",
		HasHero: true,
	}
	renderTemplate(w, "home.html", data)
}
