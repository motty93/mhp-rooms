package handlers

import (
	"context"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"mhp-rooms/internal/middleware"
)

func TestRenderTemplateUsesPublicCanonicalURL(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SITE_URL", "https://www.huntershub.net")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/login?redirect=/rooms", nil)
	renderTemplate(w, r, "login.tmpl", TemplateData{Title: "ログイン"})

	if w.Code != 200 {
		t.Fatalf("status = %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	for _, tag := range []string{
		`<link rel="canonical" href="https://www.huntershub.net/auth/login" />`,
		`<meta property="og:url" content="https://www.huntershub.net/auth/login" />`,
		`<meta name="twitter:url" content="https://www.huntershub.net/auth/login" />`,
	} {
		if !strings.Contains(body, tag) {
			t.Errorf("%q がありません", tag)
		}
	}
	if strings.Contains(body, "localhost:8080") {
		t.Error("本番URL設定時のHTMLに localhost URL が含まれています")
	}
}

func TestRenderHomeStructuredDataUsesPublicSiteURL(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SITE_URL", "https://www.huntershub.net")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	renderTemplate(w, r, "home.tmpl", TemplateData{Title: "HuntersHub"})

	if w.Code != 200 {
		t.Fatalf("status = %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "www.huntershub.net") {
		t.Error("ホームの構造化データに公開サイトURLがありません")
	}
	if strings.Contains(body, `https:\/\/huntershub.net`) {
		t.Error("ホームの構造化データにwwwなしの旧URLが残っています")
	}
}

func TestRenderTemplateExcludesAuthenticatedMarkupForGuests(t *testing.T) {
	chdirRepoRoot(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/login", nil)
	renderTemplate(w, r, "login.tmpl", TemplateData{Title: "ログイン"})

	if w.Code != 200 {
		t.Fatalf("status = %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	for _, markup := range []string{
		`href="/profile"`,
		"ログアウト",
		`@click="$store.roomCreate.open`,
		"新しい部屋を作成できません",
		"他の部屋に参加中です",
	} {
		if strings.Contains(body, markup) {
			t.Errorf("未認証HTMLに認証済みマークアップ %q が含まれています", markup)
		}
	}

	wantCopyright := "&copy; " + strconv.Itoa(time.Now().Year()) + " HuntersHub. All rights reserved."
	if !strings.Contains(body, wantCopyright) {
		t.Errorf("コピーライト = %q を含みません", wantCopyright)
	}
}

func TestRenderTemplateIncludesAuthenticatedMarkupForAuthenticatedUsers(t *testing.T) {
	chdirRepoRoot(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms", nil)
	r = r.WithContext(context.WithValue(r.Context(), middleware.UserContextKey, &middleware.AuthUser{ID: "user-id"}))
	renderTemplate(w, r, "login.tmpl", TemplateData{Title: "ログイン"})

	if w.Code != 200 {
		t.Fatalf("status = %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	for _, markup := range []string{
		`href="/profile"`,
		"ログアウト",
		`@click="$store.roomCreate.open`,
		"新しい部屋を作成できません",
	} {
		if !strings.Contains(body, markup) {
			t.Errorf("認証済みHTMLに %q が含まれていません", markup)
		}
	}
}
