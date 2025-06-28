package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func (h *PageHandler) Contact(w http.ResponseWriter, r *http.Request) {
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

func (h *PageHandler) handleContactSubmission(w http.ResponseWriter, r *http.Request) {
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

func (h *PageHandler) validateContactForm(data ContactFormData) error {
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

	if !isValidEmail(strings.TrimSpace(data.Email)) {
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