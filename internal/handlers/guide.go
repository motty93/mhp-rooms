package handlers

import "net/http"

func (h *Handler) GuideHandler(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "使い方ガイド",
	}
	renderTemplate(w, "guide.html", data)
}