package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestResolvePublicDisplayName(t *testing.T) {
	userID := uuid.MustParse("12345678-1234-5678-9abc-def012345678")
	username := "  hunter  "
	usernameWithAtMark := "  @@hunter  "
	blankUsername := "  \t"

	tests := []struct {
		name        string
		displayName string
		username    *string
		userID      uuid.UUID
		want        string
	}{
		{
			name:        "設定済みの表示名は前後空白を除いて使用する",
			displayName: "  ハンター太郎  ",
			username:    &username,
			userID:      userID,
			want:        "ハンター太郎",
		},
		{
			name:        "表示名が空白ならユーザー名を使用する",
			displayName: "  ",
			username:    &username,
			userID:      userID,
			want:        "@hunter",
		},
		{
			name:     "ユーザー名の先頭のアットマークは一つにする",
			username: &usernameWithAtMark,
			userID:   userID,
			want:     "@hunter",
		},
		{
			name:   "ユーザー名がnilならUUID先頭8文字を使用する",
			userID: userID,
			want:   "@12345678",
		},
		{
			name:     "ユーザー名が空白ならUUID先頭8文字を使用する",
			username: &blankUsername,
			userID:   userID,
			want:     "@12345678",
		},
		{
			name: "ユーザーIDもなければunknownを使用する",
			want: "@unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolvePublicDisplayName(tt.displayName, tt.username, tt.userID); got != tt.want {
				t.Errorf("ResolvePublicDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}
