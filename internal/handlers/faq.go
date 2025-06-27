package handlers

import "net/http"

func (h *Handler) FAQ(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "よくある質問",
	}
	renderTemplate(w, "faq.html", data)
}