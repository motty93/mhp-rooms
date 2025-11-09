package handlers

import (
	"encoding/xml"
	"net/http"
	"time"
)

// URL represents a single URL in the sitemap
type URL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   float64  `xml:"priority,omitempty"`
}

// URLSet represents the root element of the sitemap
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// SitemapHandler generates XML sitemap for the website
func (h *PageHandler) Sitemap(w http.ResponseWriter, r *http.Request) {
	baseURL := "https://huntershub.com"
	now := time.Now().Format("2006-01-02")

	// サイトマップのURL定義
	urls := []URL{
		{
			Loc:        baseURL + "/",
			LastMod:    now,
			ChangeFreq: "daily",
			Priority:   1.0,
		},
		{
			Loc:        baseURL + "/rooms",
			LastMod:    now,
			ChangeFreq: "always",
			Priority:   0.9,
		},
		{
			Loc:        baseURL + "/auth/login",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.8,
		},
		{
			Loc:        baseURL + "/auth/register",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.8,
		},
		{
			Loc:        baseURL + "/contact",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.7,
		},
		{
			Loc:        baseURL + "/terms",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.5,
		},
		{
			Loc:        baseURL + "/privacy",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.5,
		},
		{
			Loc:        baseURL + "/faq",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.6,
		},
		{
			Loc:        baseURL + "/guide",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.7,
		},
		{
			Loc:        baseURL + "/auth/login",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.8,
		},
		{
			Loc:        baseURL + "/auth/register",
			LastMod:    now,
			ChangeFreq: "monthly",
			Priority:   0.8,
		},
	}

	// URLSet構造体の作成
	urlSet := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	// XMLヘッダーとコンテンツタイプの設定
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))

	// XMLエンコード
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlSet); err != nil {
		http.Error(w, "Error generating sitemap", http.StatusInternalServerError)
		return
	}
}
