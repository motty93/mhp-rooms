package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"mhp-rooms/internal/repository"
)

type Handler struct {
	repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

type TemplateData struct {
	Title    string
	HasHero  bool
	User     interface{} // 将来的にユーザー情報を格納
	PageData interface{} // ページ固有のデータ
}

func renderTemplate(w http.ResponseWriter, templateName string, data TemplateData) {
	funcMap := template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}

	// 必要なテンプレートファイルを全て読み込み
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "layouts", "base.html"),
		filepath.Join("templates", "components", "header.html"),
		filepath.Join("templates", "components", "footer.html"),
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

func (h *Handler) TermsHandler(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "利用規約",
	}
	renderTemplate(w, "terms.html", data)
}

func (h *Handler) PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "プライバシーポリシー",
	}
	renderTemplate(w, "privacy.html", data)
}

func (h *Handler) ContactHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data := TemplateData{
			Title: "お問い合わせ",
		}
		renderTemplate(w, "contact.html", data)
	case "POST":
		h.handleContactSubmission(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type ContactFormData struct {
	InquiryType   string `json:"inquiryType"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Subject       string `json:"subject"`
	Message       string `json:"message"`
	PrivacyAgreed bool   `json:"privacyAgreed"`
}

func (h *Handler) handleContactSubmission(w http.ResponseWriter, r *http.Request) {
	var formData ContactFormData
	
	if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if err := h.validateContactForm(formData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("お問い合わせを受信しました: [%s] %s <%s> - %s", 
		formData.InquiryType, formData.Name, formData.Email, formData.Subject)
	
	log.Printf("お問い合わせ詳細:\n種類: %s\n名前: %s\nメール: %s\n件名: %s\n内容: %s\n時刻: %s",
		formData.InquiryType, formData.Name, formData.Email, formData.Subject, 
		formData.Message, time.Now().Format("2006-01-02 15:04:05"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "お問い合わせを受け付けました",
	})
}

func (h *Handler) validateContactForm(data ContactFormData) error {
	if data.InquiryType == "" {
		return fmt.Errorf("お問い合わせ種類を選択してください")
	}
	
	if strings.TrimSpace(data.Name) == "" {
		return fmt.Errorf("お名前を入力してください")
	}
	
	if len(strings.TrimSpace(data.Name)) > 100 {
		return fmt.Errorf("お名前は100文字以内で入力してください")
	}
	
	if strings.TrimSpace(data.Email) == "" {
		return fmt.Errorf("メールアドレスを入力してください")
	}
	
	if !h.isValidEmail(strings.TrimSpace(data.Email)) {
		return fmt.Errorf("正しいメールアドレスの形式で入力してください")
	}
	
	if strings.TrimSpace(data.Subject) == "" {
		return fmt.Errorf("件名を入力してください")
	}
	
	if len(strings.TrimSpace(data.Subject)) > 200 {
		return fmt.Errorf("件名は200文字以内で入力してください")
	}
	
	if strings.TrimSpace(data.Message) == "" {
		return fmt.Errorf("お問い合わせ内容を入力してください")
	}
	
	if len(data.Message) > 2000 {
		return fmt.Errorf("お問い合わせ内容は2000文字以内で入力してください")
	}
	
	if len(strings.TrimSpace(data.Message)) < 10 {
		return fmt.Errorf("お問い合わせ内容は10文字以上で入力してください")
	}
	
	if !data.PrivacyAgreed {
		return fmt.Errorf("プライバシーポリシーに同意してください")
	}
	
	return nil
}

func (h *Handler) isValidEmail(email string) bool {
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

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded block">
    <strong>Hello World!</strong> Go + HTMX + Tailwind CSS + Air ホットリロードで動作しています！
</div>
<script>
    document.getElementById('hello-result').classList.remove('hidden');
</script>`)
}
