package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/view"
)

type BaseHandler struct {
	repo *repository.Repository
}

type TemplateData = view.Data

func renderTemplate(w http.ResponseWriter, templateName string, data TemplateData) {
	view.Template(w, templateName, data)
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if len(email) > 254 {
		return false
	}

	if !emailRegex.MatchString(email) {
		return false
	}

	if strings.Count(email, "@") != 1 {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart := parts[0]
	domainPart := parts[1]

	if len(localPart) == 0 || len(localPart) > 64 {
		return false
	}

	if len(domainPart) == 0 || len(domainPart) > 253 {
		return false
	}

	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return false
	}

	if strings.Contains(domainPart, "..") {
		return false
	}

	return true
}

func renderPartialTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	return view.Partial(w, templateName, data)
}
