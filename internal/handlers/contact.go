package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/utils"

	"github.com/google/uuid"
)

func (h *PageHandler) Contact(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data := TemplateData{
			Title: "お問い合わせ",
		}
		renderTemplate(w, "contact.tmpl", data)
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

	// テストデータかどうかを判定
	isTestData := isTestData(formData.Name, formData.Subject, formData.Message)

	// IPアドレスを取得
	ipAddress := getClientIP(r)

	// User Agentを取得
	userAgent := r.UserAgent()

	// 認証情報を取得
	var supabaseUserID *uuid.UUID
	isAuthenticated := false
	if dbUser, ok := middleware.GetDBUserFromContext(r.Context()); ok {
		isAuthenticated = true
		supabaseUserID = &dbUser.SupabaseUserID
	}

	// Contactモデルを作成
	contact := &models.Contact{
		InquiryType:     formData.InquiryType,
		Name:            strings.TrimSpace(formData.Name),
		Email:           strings.TrimSpace(formData.Email),
		Subject:         strings.TrimSpace(formData.Subject),
		Message:         formData.Message,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		IsAuthenticated: isAuthenticated,
		SupabaseUserID:  supabaseUserID,
	}

	// テストデータでない場合のみDBに保存
	if !isTestData {
		if err := h.repo.Contact.CreateContact(contact); err != nil {
			log.Printf("お問い合わせのDB保存に失敗しました: %v", err)
			// DB保存失敗してもユーザーにはエラーを返さない（Discord通知は試行）
		} else {
			log.Printf("お問い合わせをDBに保存しました: ID=%s", contact.ID)
		}
	} else {
		log.Printf("テストデータのためDB保存をスキップしました")
	}

	// Discord通知を送信（テストデータも通知）
	contactInfo := &utils.ContactInfo{
		InquiryType:     contact.InquiryType,
		Name:            contact.Name,
		Email:           contact.Email,
		Subject:         contact.Subject,
		Message:         contact.Message,
		IPAddress:       contact.IPAddress,
		UserAgent:       contact.UserAgent,
		IsAuthenticated: contact.IsAuthenticated,
		SupabaseUserID:  contact.SupabaseUserID,
	}
	if err := utils.SendContactNotificationToDiscord(
		config.AppConfig.Discord.WebhookURL,
		contactInfo,
		isTestData,
	); err != nil {
		log.Printf("Discord通知に失敗しました: %v", err)
		// Discord通知失敗してもユーザーにはエラーを返さない
	}

	log.Printf("お問い合わせを受信しました: [%s] %s <%s> - %s (テストデータ: %v)",
		formData.InquiryType, formData.Name, formData.Email, formData.Subject, isTestData)

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

// isTestData テストデータかどうかを判定
func isTestData(name, subject, message string) bool {
	testKeywords := []string{"test", "テスト", "てすと", "TEST"}

	// お名前、件名、お問い合わせ内容のいずれかに含まれていればテストデータとみなす
	for _, keyword := range testKeywords {
		if strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(subject), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(message), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// getClientIP クライアントのIPアドレスを取得
func getClientIP(r *http.Request) string {
	// X-Forwarded-Forヘッダーを優先（プロキシ経由の場合）
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// 複数のIPがある場合は最初のものを使用
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// X-Real-IPヘッダーをチェック
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// それ以外の場合はRemoteAddrを使用
	ip := r.RemoteAddr
	// ポート番号を削除
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}
