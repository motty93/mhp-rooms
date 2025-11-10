package handlers

import (
	"fmt"

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
