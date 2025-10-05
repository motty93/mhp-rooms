package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ContactInfo お問合せ情報（インポートサイクル回避のため）
type ContactInfo struct {
	InquiryType     string
	Name            string
	Email           string
	Subject         string
	Message         string
	IPAddress       string
	UserAgent       string
	IsAuthenticated bool
	SupabaseUserID  *uuid.UUID
}

// DiscordWebhook Discord Webhook通知の構造体
type DiscordWebhook struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed Discord Embedの構造体
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
}

// DiscordEmbedField Discord Embed Fieldの構造体
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// SendContactNotificationToDiscord お問合せ内容をDiscordに通知
func SendContactNotificationToDiscord(webhookURL string, contact *ContactInfo, isTestData bool) error {
	if webhookURL == "" {
		log.Println("Discord Webhook URLが設定されていません。通知をスキップします。")
		return nil
	}

	// タイトルを設定
	title := "📧 新しいお問い合わせ"
	if isTestData {
		title = "🧪 テストデータ（DB保存なし）"
	}

	// 認証情報を整形
	authInfo := "未ログイン"
	if contact.IsAuthenticated && contact.SupabaseUserID != nil {
		authInfo = fmt.Sprintf("認証済み（ID: %s）", contact.SupabaseUserID.String())
	}

	// Embedを作成
	embed := DiscordEmbed{
		Title: title,
		Color: getColorByInquiryType(contact.InquiryType, isTestData),
		Fields: []DiscordEmbedField{
			{Name: "種類", Value: contact.InquiryType, Inline: true},
			{Name: "お名前", Value: contact.Name, Inline: true},
			{Name: "メールアドレス", Value: contact.Email, Inline: false},
			{Name: "件名", Value: contact.Subject, Inline: false},
			{Name: "お問い合わせ内容", Value: truncateMessage(contact.Message, 1000), Inline: false},
			{Name: "認証情報", Value: authInfo, Inline: true},
			{Name: "IPアドレス", Value: contact.IPAddress, Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// User Agentが長い場合は省略
	if contact.UserAgent != "" {
		embed.Fields = append(embed.Fields, DiscordEmbedField{
			Name:   "User Agent",
			Value:  truncateMessage(contact.UserAgent, 200),
			Inline: false,
		})
	}

	webhook := DiscordWebhook{
		Embeds: []DiscordEmbed{embed},
	}

	// JSONに変換
	payload, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("Discord Webhook payloadの作成に失敗しました: %w", err)
	}

	// HTTPリクエストを送信
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("Discord Webhookの送信に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Discord Webhookがエラーを返しました: %d", resp.StatusCode)
	}

	log.Println("Discord通知を送信しました")
	return nil
}

// getColorByInquiryType お問合せ種類に応じた色を返す
func getColorByInquiryType(inquiryType string, isTestData bool) int {
	if isTestData {
		return 0x808080 // グレー（テストデータ）
	}

	switch inquiryType {
	case "バグ報告":
		return 0xFF0000 // 赤
	case "機能要望":
		return 0x00FF00 // 緑
	case "使い方・操作方法":
		return 0x0000FF // 青
	case "アカウント関連":
		return 0xFFFF00 // 黄色
	case "その他":
		return 0x808080 // グレー
	default:
		return 0x808080 // グレー
	}
}

// truncateMessage メッセージを指定文字数で切り詰める
func truncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength] + "..."
}
