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
	SSEHost  string // SSEサーバーのホスト
}

// テンプレート関数の定義

// toLower は文字列を小文字に変換する
func toLower(s string) string {
	return strings.ToLower(s)
}

// toJSON は値をJSONに変換してテンプレートで使用可能にする
func toJSON(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(b)
}

// makeMap キーと値のペアからマップを作成する（テンプレート内で動的にデータを構築する際に使用）
func makeMap(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("makeMap called with odd number of arguments")
	}
	m := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("makeMap keys must be strings")
		}
		m[key] = values[i+1]
	}
	return m, nil
}

// ゲームバージョンアイコンをHTMLとして返す
func gameVersionIconHTML(code string) template.HTML {
	return template.HTML(utils.GetGameVersionIcon(code))
}

// ポインタ文字列を通常の文字列に変換（nil安全）
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ポインタ文字列を通常の文字列に変換（旧名、互換性のため維持）
func derefString(s *string) string {
	return safeString(s)
}

// ポインタ文字列が値を持つかチェック
func hasStringValue(s *string) bool {
	return s != nil && *s != ""
}

// JavaScript文字列内で使用するための文字列エスケープ
func jsEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// ポインタ文字列に対してJavaScriptエスケープ
func jsEscapePtr(s *string) string {
	if s == nil {
		return ""
	}
	return jsEscape(*s)
}

func getCommonFuncMap() template.FuncMap {
	return template.FuncMap{
		"lower":            toLower,
		"json":             toJSON,
		"map":              makeMap,
		"gameVersionColor": utils.GetGameVersionColor,
		"gameVersionIcon":  gameVersionIconHTML,
		"gameVersionAbbr":  utils.GetGameVersionAbbreviation,
		"stringPtr":        derefString, // 互換性のため維持
		"safeString":       safeString,  // nil安全な文字列変換（推奨）
		"hasStringValue":   hasStringValue,
		"jsEscape":         jsEscape,
		"jsEscapePtr":      jsEscapePtr,
	}
}

// renderTemplate は全てのハンドラーで使用可能なテンプレートレンダリング関数
func renderTemplate(w http.ResponseWriter, templateName string, data TemplateData) {
	funcMap := getCommonFuncMap()

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
	funcMap := getCommonFuncMap()

	// テンプレート名からファイル名を生成
	templateFileName := templateName + ".tmpl"
	templateFiles := []string{filepath.Join("templates", "components", templateFileName)}

	// profile_card_contentの場合は、依存するテンプレートも読み込む
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
	// テンプレート名でテンプレートを実行
	err = tmpl.ExecuteTemplate(w, templateName, data)
	if err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}
	return nil
}
