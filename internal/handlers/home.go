package handlers

import "net/http"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title:   "ホーム",
		HasHero: true, // ホームページはヒーローセクションがある
	}
	renderTemplate(w, "home.html", data)
}
