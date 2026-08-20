package handlers

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"mhp-rooms/internal/info"
	"mhp-rooms/internal/models"
)

func TestBuildNotificationOverview(t *testing.T) {
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	readSince := now.Add(-72 * time.Hour) // 3日前に更新情報を既読にした
	readAt := now.Add(-time.Hour)
	body := "本文"

	personal := []models.Notification{
		{BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: now.Add(-30 * time.Minute)}, Type: models.NotificationFollow, Title: "Aさんにフォローされました"},                          // 未読
		{BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: now.Add(-5 * 24 * time.Hour)}, Type: models.NotificationRoomKicked, Title: "退出", Body: &body, ReadAt: &readAt}, // 既読
	}
	updated := now.Add(-24 * time.Hour)
	articles := info.ArticleList{
		{Title: "8月のアップデート", Slug: "aug", Date: now.Add(-48 * time.Hour), Updated: &updated, Category: info.ArticleTypeRelease}, // updated が既読日時より新しい → 未読
		{Title: "古いお知らせ", Slug: "old", Date: now.Add(-10 * 24 * time.Hour), Category: info.ArticleTypeNews},                     // 既読扱い
		{Title: "下書き", Slug: "draft", Date: now, Category: info.ArticleTypeNews, Draft: true},                                   // 除外
		{Title: "ロードマップ", Slug: "roadmap", Date: now, Category: info.ArticleTypeRoadmap},                                        // カテゴリ対象外
	}

	got := buildNotificationOverview(personal, 1, articles, readSince, now)

	if got.UnreadCount != 2 {
		t.Errorf("UnreadCount = %d, want 2（個人宛1 + 更新情報1）", got.UnreadCount)
	}
	if len(got.Items) != 4 {
		t.Fatalf("Items = %d 件, want 4（個人宛2 + 更新情報2。下書きとロードマップは除外）", len(got.Items))
	}

	// 新しい順: フォロー(30分前) → 8月のアップデート(updated 1日前) → 退出(5日前) → 古いお知らせ(10日前)
	wantOrder := []string{"Aさんにフォローされました", "8月のアップデート", "退出", "古いお知らせ"}
	for i, want := range wantOrder {
		if got.Items[i].Title != want {
			t.Errorf("Items[%d].Title = %q, want %q", i, got.Items[i].Title, want)
		}
	}

	byTitle := map[string]NotificationItem{}
	for _, item := range got.Items {
		byTitle[item.Title] = item
	}
	if !byTitle["Aさんにフォローされました"].Unread || byTitle["退出"].Unread {
		t.Errorf("個人宛の未読判定が誤り: %+v / %+v", byTitle["Aさんにフォローされました"], byTitle["退出"])
	}
	if !byTitle["8月のアップデート"].Unread || byTitle["古いお知らせ"].Unread {
		t.Errorf("更新情報の未読判定が誤り: %+v / %+v", byTitle["8月のアップデート"], byTitle["古いお知らせ"])
	}
	if byTitle["8月のアップデート"].LinkURL != "/info/aug" || byTitle["8月のアップデート"].Kind != "info" {
		t.Errorf("更新情報のリンク/種別が誤り: %+v", byTitle["8月のアップデート"])
	}
	if byTitle["退出"].Body != "本文" {
		t.Errorf("Body が引き継がれていない: %+v", byTitle["退出"])
	}
}

func TestBuildNotificationOverviewLimits(t *testing.T) {
	now := time.Now()
	var personal []models.Notification
	for i := 0; i < notificationPanelLimit+5; i++ {
		personal = append(personal, models.Notification{BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: now.Add(-time.Duration(i) * time.Minute)}, Type: models.NotificationFollow, Title: "n"})
	}
	var articles info.ArticleList
	for i := 0; i < notificationInfoLimit+3; i++ {
		articles = append(articles, &info.Article{Title: "i", Slug: "s", Date: now.Add(-time.Duration(i) * time.Hour), Category: info.ArticleTypeNews})
	}

	got := buildNotificationOverview(personal, int64(len(personal)), articles, now.Add(-24*time.Hour), now)

	if len(got.Items) != notificationPanelLimit {
		t.Errorf("Items = %d 件, want %d（上限で打ち切り）", len(got.Items), notificationPanelLimit)
	}
	// 個人宛の未読数は一覧の上限に関係なく総数、更新情報は上位 notificationInfoLimit 件まで
	if got.UnreadCount != len(personal)+notificationInfoLimit {
		t.Errorf("UnreadCount = %d, want %d", got.UnreadCount, len(personal)+notificationInfoLimit)
	}
}
