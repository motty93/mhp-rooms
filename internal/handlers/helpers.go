package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"mhp-rooms/internal/config"

	"github.com/google/uuid"
)

// BuildOGPImageURL OGP画像URLを生成
func BuildOGPImageURL(roomID uuid.UUID, ogVersion int) string {
	ogBucket := config.GetEnv("OG_BUCKET", "")
	ogPrefix := config.GetEnv("OG_PREFIX", "dev")

	if ogBucket != "" {
		return fmt.Sprintf(
			"https://storage.googleapis.com/%s/%s/ogp/rooms/%s.png?v=%d",
			ogBucket, ogPrefix, roomID, ogVersion,
		)
	}

	siteURL := config.GetEnv("SITE_URL", "http://localhost:8080")
	return fmt.Sprintf(
		"%s/tmp/images/%s/ogp/rooms/%s.png?v=%d",
		siteURL, ogPrefix, roomID, ogVersion,
	)
}

// withCanonicalURL ensures every view receives a canonical URL derived from the request path.
func withCanonicalURL(r *http.Request, data TemplateData) TemplateData {
	if r == nil || data.CanonicalURL != "" {
		return data
	}
	data.CanonicalURL = buildCanonicalURL(r)
	return data
}

func buildCanonicalURL(r *http.Request) string {
	base := strings.TrimRight(config.GetEnv("SITE_URL", "http://localhost:8080"), "/")
	path := r.URL.Path
	if path == "" {
		path = "/"
	}
	return base + path
}
