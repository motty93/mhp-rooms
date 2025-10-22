package view

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type Data struct {
	Title    string
	HasHero  bool
	User     interface{}
	PageData interface{}
	SSEHost  string
}

func Template(w http.ResponseWriter, templateName string, data Data) {
	funcMap := TemplateFuncs()

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "layouts", "base.tmpl"),
		filepath.Join("templates", "components", "header.tmpl"),
		filepath.Join("templates", "components", "footer.tmpl"),
		filepath.Join("templates", "components", "room_create_button.tmpl"),
		filepath.Join("templates", "components", "room_create_modal.tmpl"),
		filepath.Join("templates", "components", "profile_view.tmpl"),
		filepath.Join("templates", "components", "profile_edit_form.tmpl"),
		filepath.Join("templates", "components", "profile_activity.tmpl"),
		filepath.Join("templates", "components", "profile_rooms.tmpl"),
		filepath.Join("templates", "components", "profile_followers.tmpl"),
		filepath.Join("templates", "components", "profile_following.tmpl"),
		filepath.Join("templates", "components", "follow_buttons.tmpl"),
		filepath.Join("templates", "components", "block_report_buttons.tmpl"),
		filepath.Join("templates", "components", "report_modal.tmpl"),
		filepath.Join("templates", "pages", templateName),
	)
	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func Partial(w http.ResponseWriter, templateName string, data interface{}) error {
	funcMap := TemplateFuncs()

	templateFileName := templateName + ".tmpl"
	templateFiles := []string{filepath.Join("templates", "components", templateFileName)}

	if templateName == "profile_card_content" {
		templateFiles = append(templateFiles,
			filepath.Join("templates", "components", "follow_buttons.tmpl"),
			filepath.Join("templates", "components", "block_report_buttons.tmpl"),
		)
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(templateFiles...)
	if err != nil {
		return fmt.Errorf("template parsing error: %w", err)
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, templateName, data)
	if err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}
	return nil
}
