package utils

import (
	"strings"

	"github.com/google/uuid"
)

// ResolvePublicDisplayName は公開画面に表示する安全なユーザー名を返します。
func ResolvePublicDisplayName(displayName string, username *string, userID uuid.UUID) string {
	if name := strings.TrimSpace(displayName); name != "" {
		return name
	}

	if username != nil {
		name := strings.TrimLeft(strings.TrimSpace(*username), "@")
		if name != "" {
			return "@" + name
		}
	}

	if userID != uuid.Nil {
		return "@" + userID.String()[:8]
	}

	return "@unknown"
}
