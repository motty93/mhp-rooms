package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/utils"
)

// BaseHandler は全てのハンドラーが共通で使用する構造体
type BaseHandler struct {
	repo *repository.Repository
}

type TemplateData struct {
	Title    string
	HasHero  bool
	User     interface{}
	PageData interface{}
}

// renderTemplate は全てのハンドラーで使用可能なテンプレートレンダリング関数
func renderTemplate(w http.ResponseWriter, templateName string, data TemplateData) {
	funcMap := template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"json": func(v interface{}) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("[]")
			}
			return template.JS(b)
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict called with odd number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"gameVersionColor": utils.GetGameVersionColor,
		"gameVersionIcon": func(code string) template.HTML {
			return template.HTML(utils.GetGameVersionIcon(code))
		},
		"gameVersionAbbr": utils.GetGameVersionAbbreviation,
		"stringPtr": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"hasStringValue": func(s *string) bool {
			return s != nil && *s != ""
		},
	}

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

// isValidEmail はメールアドレスの妥当性を検証する共通関数
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

// renderPartialTemplate は部分テンプレート（コンポーネント）をレンダリングする関数
func renderPartialTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	funcMap := template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"json": func(v interface{}) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("[]")
			}
			return template.JS(b)
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict called with odd number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"gameVersionColor": utils.GetGameVersionColor,
		"gameVersionIcon": func(code string) template.HTML {
			return template.HTML(utils.GetGameVersionIcon(code))
		},
		"gameVersionAbbr": utils.GetGameVersionAbbreviation,
		"stringPtr": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"hasStringValue": func(s *string) bool {
			return s != nil && *s != ""
		},
	}

	// プロフィール関連のテンプレートの場合は、関連するテンプレートも一緒に読み込む
	templateFiles := []string{filepath.Join("templates", "components", templateName)}
	if strings.HasPrefix(templateName, "profile_") {
		templateFiles = append(templateFiles,
			filepath.Join("templates", "components", "profile_view.tmpl"),
			filepath.Join("templates", "components", "profile_edit_form.tmpl"),
		)
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(templateFiles...)
	if err != nil {
		return fmt.Errorf("template parsing error: %w", err)
	}

	w.Header().Set("Content-Type", "text/html")
	// ".tmpl"を除去してテンプレート名を取得
	templateBaseName := templateName[:len(templateName)-5]
	err = tmpl.ExecuteTemplate(w, templateBaseName, data)
	if err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}
	return nil
}
