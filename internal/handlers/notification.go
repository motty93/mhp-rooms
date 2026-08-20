package handlers

import (
	"log"
	"net/http"
	"sort"
	"time"

	"mhp-rooms/internal/info"
	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

const (
	notificationPanelLimit = 20 // お知らせパネルに表示する最大件数
	notificationInfoLimit  = 5  // パネルに載せる更新情報の最大件数
)

// NotificationItem お知らせパネルの 1 行（個人宛のお知らせと更新情報を同じ形にそろえる）
type NotificationItem struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"` // personal / info
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	LinkURL   string    `json:"link_url"`
	CreatedAt time.Time `json:"created_at"`
	TimeAgo   string    `json:"time_ago"`
	Unread    bool      `json:"unread"`
}

// NotificationOverview お知らせパネル用のまとめ
type NotificationOverview struct {
	UnreadCount int                `json:"unread_count"`
	Items       []NotificationItem `json:"items"`
}

type NotificationHandler struct {
	BaseHandler
	logger       *log.Logger
	articlesPath string
	generator    *info.Generator
}

func NewNotificationHandler(repo *repository.Repository, generator *info.Generator) *NotificationHandler {
	return &NotificationHandler{
		BaseHandler:  BaseHandler{repo: repo},
		logger:       log.New(log.Writer(), "[NotificationHandler] ", log.LstdFlags),
		articlesPath: "static/generated/info/articles.json",
		generator:    generator,
	}
}

// List 未読数と最新のお知らせ（個人宛 + 更新情報）を JSON で返す
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていません", http.StatusUnauthorized)
		return
	}

	personal, err := h.repo.Notification.ListByUser(dbUser.ID, notificationPanelLimit)
	if err != nil {
		h.logger.Printf("お知らせ取得エラー: %v", err)
		http.Error(w, "お知らせの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	personalUnread, err := h.repo.Notification.CountUnread(dbUser.ID)
	if err != nil {
		h.logger.Printf("未読数取得エラー: %v", err)
		http.Error(w, "お知らせの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 更新情報は「最後に既読にした日時」（未設定なら登録日時）より新しいものを未読として扱う
	infoReadSince := dbUser.CreatedAt
	state, err := h.repo.Notification.GetState(dbUser.ID)
	if err != nil {
		h.logger.Printf("閲覧状態取得エラー: %v", err)
	} else if state != nil && state.InfoReadAt != nil {
		infoReadSince = *state.InfoReadAt
	}

	// 更新情報が読めなくてもお知らせ自体は返す
	articles, err := loadArticlesWithFallback(h.articlesPath, h.generator)
	if err != nil {
		h.logger.Printf("更新情報の読み込みエラー: %v", err)
		articles = nil
	}

	respondWithJSON(w, http.StatusOK, buildNotificationOverview(personal, personalUnread, articles, infoReadSince, time.Now()))
}

// MarkAllRead 個人宛のお知らせと更新情報をすべて既読にする
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	dbUser, exists := middleware.GetDBUserFromContext(r.Context())
	if !exists || dbUser == nil {
		http.Error(w, "認証されていません", http.StatusUnauthorized)
		return
	}

	now := time.Now()
	if err := h.repo.Notification.MarkAllRead(dbUser.ID, now); err != nil {
		h.logger.Printf("既読更新エラー: %v", err)
		http.Error(w, "既読の更新に失敗しました", http.StatusInternalServerError)
		return
	}
	if err := h.repo.Notification.UpsertInfoReadAt(dbUser.ID, now); err != nil {
		h.logger.Printf("更新情報の既読更新エラー: %v", err)
		http.Error(w, "既読の更新に失敗しました", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"unread_count": 0})
}

// buildNotificationOverview 個人宛のお知らせと更新情報を新しい順にマージし、未読数を数える。
// personalUnread は一覧の件数上限に関係なく数えた個人宛の未読数、infoReadSince より新しい更新情報を未読とする
func buildNotificationOverview(personal []models.Notification, personalUnread int64, articles info.ArticleList, infoReadSince, now time.Time) NotificationOverview {
	items := make([]NotificationItem, 0, len(personal)+notificationInfoLimit)

	for _, n := range personal {
		items = append(items, NotificationItem{
			ID:        n.ID.String(),
			Kind:      "personal",
			Type:      n.Type,
			Title:     n.Title,
			Body:      getStringValue(n.Body),
			LinkURL:   getStringValue(n.LinkURL),
			CreatedAt: n.CreatedAt,
			TimeAgo:   formatRelativeTime(n.CreatedAt),
			Unread:    n.IsUnread(),
		})
	}

	infoUnread := 0
	for _, article := range latestInfoArticles(articles, notificationInfoLimit) {
		publishedAt := article.Date
		if article.Updated != nil && article.Updated.After(publishedAt) {
			publishedAt = *article.Updated
		}
		unread := publishedAt.After(infoReadSince)
		if unread {
			infoUnread++
		}
		items = append(items, NotificationItem{
			ID:        "info:" + article.Slug,
			Kind:      "info",
			Type:      "info",
			Title:     article.Title,
			Body:      article.Summary,
			LinkURL:   "/info/" + article.Slug,
			CreatedAt: publishedAt,
			TimeAgo:   formatRelativeTime(publishedAt),
			Unread:    unread,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if len(items) > notificationPanelLimit {
		items = items[:notificationPanelLimit]
	}

	return NotificationOverview{
		UnreadCount: int(personalUnread) + infoUnread,
		Items:       items,
	}
}

// latestInfoArticles お知らせに載せる更新情報（リリース・ニュース・メンテナンス）を新しい順に limit 件返す
func latestInfoArticles(articles info.ArticleList, limit int) info.ArticleList {
	if len(articles) == 0 {
		return nil
	}

	var filtered info.ArticleList
	for _, category := range []info.ArticleType{info.ArticleTypeRelease, info.ArticleTypeNews, info.ArticleTypeMaintenance} {
		filtered = append(filtered, articles.FilterByCategory(category)...)
	}
	filtered = filtered.ExcludeDrafts().SortByDateDesc()
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}
