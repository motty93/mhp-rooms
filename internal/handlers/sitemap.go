package handlers

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/info"
)

const (
	sitemapDateFormat = "2006-01-02"
	sitemapRoomLimit  = 100
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
	baseURL := strings.TrimRight(config.GetEnv("SITE_URL", "http://localhost:8080"), "/")
	now := time.Now()

	urls := make([]URL, 0, 32)
	urls = append(urls, defaultSitemapEntries(baseURL, now)...)

	if articleURLs, err := h.buildArticleURLs(baseURL); err != nil {
		log.Printf("sitemap: failed to load articles: %v", err)
	} else {
		urls = append(urls, articleURLs...)
	}

	if blogURLs, err := h.buildBlogURLs(baseURL); err != nil {
		log.Printf("sitemap: failed to load blog articles: %v", err)
	} else {
		urls = append(urls, blogURLs...)
	}

	if roomURLs, err := h.buildRoomURLs(baseURL); err != nil {
		log.Printf("sitemap: failed to load rooms: %v", err)
	} else {
		urls = append(urls, roomURLs...)
	}

	urlSet := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlSet); err != nil {
		http.Error(w, "Error generating sitemap", http.StatusInternalServerError)
		return
	}
}

func defaultSitemapEntries(baseURL string, now time.Time) []URL {
	lastMod := formatSitemapDate(now)
	return []URL{
		{
			Loc:        baseURL + "/",
			LastMod:    lastMod,
			ChangeFreq: "daily",
			Priority:   1.0,
		},
		{
			Loc:        baseURL + "/rooms",
			LastMod:    lastMod,
			ChangeFreq: "hourly",
			Priority:   0.9,
		},
		{
			Loc:        baseURL + "/contact",
			LastMod:    lastMod,
			ChangeFreq: "monthly",
			Priority:   0.6,
		},
		{
			Loc:        baseURL + "/auth/login",
			LastMod:    lastMod,
			ChangeFreq: "monthly",
			Priority:   0.5,
		},
		{
			Loc:        baseURL + "/auth/register",
			LastMod:    lastMod,
			ChangeFreq: "monthly",
			Priority:   0.5,
		},
	}
}

func (h *PageHandler) buildArticleURLs(baseURL string) ([]URL, error) {
	articles, err := h.loadArticles()
	if err != nil {
		return nil, err
	}

	urls := make([]URL, 0, len(articles)+6)
	staticCategories := []struct {
		category   info.ArticleType
		path       string
		changeFreq string
		priority   float64
	}{
		{info.ArticleTypeGuide, "/guide", "monthly", 0.7},
		{info.ArticleTypeFAQ, "/faq", "monthly", 0.6},
		{info.ArticleTypeTerms, "/terms", "yearly", 0.5},
		{info.ArticleTypePrivacy, "/privacy", "yearly", 0.5},
		{info.ArticleTypeOperator, "/operator", "yearly", 0.4},
	}

	for _, entry := range staticCategories {
		article := firstArticleByCategory(articles, entry.category)
		if article == nil {
			continue
		}
		urls = append(urls, URL{
			Loc:        baseURL + entry.path,
			LastMod:    formatSitemapDate(articleTimestamp(article)),
			ChangeFreq: entry.changeFreq,
			Priority:   entry.priority,
		})
	}

	if latest := newestArticleDate(articles, info.ArticleTypeNews, info.ArticleTypeRelease, info.ArticleTypeMaintenance, info.ArticleTypeFeature); latest != nil {
		urls = append(urls, URL{
			Loc:        baseURL + "/info",
			LastMod:    formatSitemapDate(*latest),
			ChangeFreq: "daily",
			Priority:   0.7,
		})
	}

	if latest := newestArticleDate(articles, info.ArticleTypeRoadmap); latest != nil {
		urls = append(urls, URL{
			Loc:        baseURL + "/roadmap",
			LastMod:    formatSitemapDate(*latest),
			ChangeFreq: "weekly",
			Priority:   0.6,
		})
	}

	for _, article := range articles {
		if !shouldExposeInfoArticle(article.Category) {
			continue
		}
		urls = append(urls, URL{
			Loc:        fmt.Sprintf("%s/info/%s", baseURL, article.Slug),
			LastMod:    formatSitemapDate(articleTimestamp(article)),
			ChangeFreq: "weekly",
			Priority:   0.6,
		})
	}

	return urls, nil
}

func (h *PageHandler) buildBlogURLs(baseURL string) ([]URL, error) {
	articles, err := h.loadArticles()
	if err != nil {
		return nil, err
	}

	// ブログ一覧ページ
	urls := make([]URL, 0, len(articles)+1)
	if latest := newestArticleDate(articles, info.ArticleTypeBlogCommunity, info.ArticleTypeBlogTechnical, info.ArticleTypeBlogTroubleshooting); latest != nil {
		urls = append(urls, URL{
			Loc:        baseURL + "/blog",
			LastMod:    formatSitemapDate(*latest),
			ChangeFreq: "weekly",
			Priority:   0.8,
		})
	}

	// 各ブログ記事
	for _, article := range articles {
		if !isBlogArticle(article.Category) {
			continue
		}
		urls = append(urls, URL{
			Loc:        fmt.Sprintf("%s/blog/%s", baseURL, article.Slug),
			LastMod:    formatSitemapDate(articleTimestamp(article)),
			ChangeFreq: "monthly",
			Priority:   0.7,
		})
	}

	return urls, nil
}

func isBlogArticle(category info.ArticleType) bool {
	switch category {
	case info.ArticleTypeBlogCommunity,
		info.ArticleTypeBlogTechnical,
		info.ArticleTypeBlogTroubleshooting:
		return true
	default:
		return false
	}
}

func (h *PageHandler) buildRoomURLs(baseURL string) ([]URL, error) {
	rooms, err := h.repo.Room.GetActiveRooms(nil, sitemapRoomLimit, 0)
	if err != nil {
		return nil, err
	}

	urls := make([]URL, 0, len(rooms)*2)
	for _, room := range rooms {
		lastMod := formatSitemapDate(room.UpdatedAt)
		urls = append(urls, URL{
			Loc:        fmt.Sprintf("%s/rooms/%s", baseURL, room.ID),
			LastMod:    lastMod,
			ChangeFreq: "hourly",
			Priority:   0.6,
		})
		urls = append(urls, URL{
			Loc:        fmt.Sprintf("%s/rooms/%s/join", baseURL, room.ID),
			LastMod:    lastMod,
			ChangeFreq: "hourly",
			Priority:   0.5,
		})
	}

	return urls, nil
}

func formatSitemapDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(sitemapDateFormat)
}

func newestArticleDate(articles info.ArticleList, categories ...info.ArticleType) *time.Time {
	var latest *time.Time
	categorySet := make(map[info.ArticleType]struct{}, len(categories))
	for _, c := range categories {
		categorySet[c] = struct{}{}
	}
	for _, article := range articles {
		if len(categorySet) > 0 {
			if _, ok := categorySet[article.Category]; !ok {
				continue
			}
		}
		timestamp := articleTimestamp(article)
		if timestamp.IsZero() {
			continue
		}
		if latest == nil || timestamp.After(*latest) {
			copy := timestamp
			latest = &copy
		}
	}
	return latest
}

func articleTimestamp(article *info.Article) time.Time {
	if article == nil {
		return time.Time{}
	}
	if article.Updated != nil {
		return *article.Updated
	}
	return article.Date
}

func firstArticleByCategory(articles info.ArticleList, category info.ArticleType) *info.Article {
	filtered := articles.FilterByCategory(category)
	if len(filtered) == 0 {
		return nil
	}
	return filtered[0]
}

func shouldExposeInfoArticle(category info.ArticleType) bool {
	switch category {
	case info.ArticleTypeNews,
		info.ArticleTypeRelease,
		info.ArticleTypeMaintenance,
		info.ArticleTypeFeature:
		return true
	default:
		return false
	}
}
