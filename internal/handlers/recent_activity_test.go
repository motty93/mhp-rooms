package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mhp-rooms/internal/models"

	"github.com/google/uuid"
)

func TestStartOfCurrentWeek(t *testing.T) {
	now := time.Date(2026, time.August, 21, 15, 30, 0, 0, time.UTC)
	got := startOfCurrentWeek(now)
	want := time.Date(2026, time.August, 17, 0, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	if !got.Equal(want) {
		t.Fatalf("週初め = %v, want %v", got, want)
	}
}

func TestRenderRecentActivityFeedPartial(t *testing.T) {
	chdirRepoRoot(t)
	w := httptest.NewRecorder()
	data := RecentActivityFeedData{Activities: []RecentActivityItem{{
		UserID:      uuid.New(),
		DisplayName: "テスト",
		Title:       "部屋を作成",
		Type:        "room_create",
		TimeAgo:     "1分前",
	}}}
	if err := renderPartialTemplate(w, "recent_activity_feed", data); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || w.Body.Len() == 0 {
		t.Fatalf("フィード描画に失敗: %d", w.Code)
	}
}

func TestRenderRecentActivityFeedUsesFallbackDisplayName(t *testing.T) {
	chdirRepoRoot(t)

	w := httptest.NewRecorder()
	data := RecentActivityFeedData{Activities: []RecentActivityItem{{
		UserID:      uuid.New(),
		DisplayName: "@fallback-user",
		Title:       "部屋を作成",
		Type:        "room_create",
		TimeAgo:     "1分前",
	}}}
	if err := renderPartialTemplate(w, "recent_activity_feed", data); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.Body.String(), ">@fallback-user</a") {
		t.Errorf("フォールバック表示名が描画されていない:\n%s", w.Body.String())
	}
}

func TestRecentActivityItemUsesUsernameWhenDisplayNameIsEmpty(t *testing.T) {
	userID := uuid.MustParse("12345678-1234-5678-9abc-def012345678")
	username := "  fallback-user  "
	activity := models.UserActivity{
		User: models.User{
			BaseModel:   models.BaseModel{ID: userID},
			DisplayName: "  ",
			Username:    &username,
		},
	}

	if got := recentActivityItem(activity).DisplayName; got != "@fallback-user" {
		t.Errorf("recentActivityItem().DisplayName = %q, want %q", got, "@fallback-user")
	}
}
