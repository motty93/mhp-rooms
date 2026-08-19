package handlers

import (
	"testing"
	"time"

	"mhp-rooms/internal/models"
)

func TestRoomToSummaryStatus(t *testing.T) {
	now := time.Now()
	host := models.DismissReasonHost
	inactive := models.DismissReasonInactive

	tests := []struct {
		name          string
		room          models.Room
		wantStatus    string
		wantNote      bool
		wantClickable bool
	}{
		{
			name:          "募集中",
			room:          models.Room{IsActive: true},
			wantStatus:    "アクティブ",
			wantClickable: true,
		},
		{
			name:       "締め切り",
			room:       models.Room{IsActive: true, IsClosed: true},
			wantStatus: "終了",
		},
		{
			name:       "ホストが解散",
			room:       models.Room{IsActive: false, DismissedAt: &now, DismissReason: &host},
			wantStatus: "削除済み",
		},
		{
			name:       "理由なしの解散（既存データ）",
			room:       models.Room{IsActive: false},
			wantStatus: "削除済み",
		},
		{
			name:       "一定期間活動がなく自動解散",
			room:       models.Room{IsActive: false, DismissedAt: &now, DismissReason: &inactive},
			wantStatus: "自動削除",
			wantNote:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roomToSummary(tt.room)
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if (got.StatusNote != "") != tt.wantNote {
				t.Errorf("StatusNote = %q, wantNote %v", got.StatusNote, tt.wantNote)
			}
			if got.IsClickable != tt.wantClickable {
				t.Errorf("IsClickable = %v, want %v", got.IsClickable, tt.wantClickable)
			}
		})
	}
}
