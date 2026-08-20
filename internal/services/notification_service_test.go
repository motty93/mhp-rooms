package services

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

// fakeNotificationRepo 作成されたお知らせを記録するだけのテスト用リポジトリ
type fakeNotificationRepo struct {
	created []*models.Notification
}

func (f *fakeNotificationRepo) Create(n *models.Notification) error {
	f.created = append(f.created, n)
	return nil
}
func (f *fakeNotificationRepo) ListByUser(uuid.UUID, int) ([]models.Notification, error) {
	return nil, nil
}
func (f *fakeNotificationRepo) CountUnread(uuid.UUID) (int64, error)   { return 0, nil }
func (f *fakeNotificationRepo) MarkAllRead(uuid.UUID, time.Time) error { return nil }
func (f *fakeNotificationRepo) GetState(uuid.UUID) (*models.UserNotificationState, error) {
	return nil, nil
}
func (f *fakeNotificationRepo) UpsertInfoReadAt(uuid.UUID, time.Time) error { return nil }

func TestNotifyRoomDismissedToMembersSkipsHost(t *testing.T) {
	fake := &fakeNotificationRepo{}
	svc := NewNotificationService(&repository.Repository{Notification: fake})

	host := uuid.New()
	guest1, guest2 := uuid.New(), uuid.New()
	room := &models.Room{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "テスト部屋", HostUserID: host}
	members := []models.RoomMember{
		{UserID: host, IsHost: true},
		{UserID: guest1},
		{UserID: guest2},
	}

	if err := svc.NotifyRoomDismissedToMembers(room, members); err != nil {
		t.Fatalf("NotifyRoomDismissedToMembers() error = %v", err)
	}

	if len(fake.created) != 2 {
		t.Fatalf("作成されたお知らせ = %d 件, want 2（ホストを除くメンバー分）", len(fake.created))
	}
	for _, n := range fake.created {
		if n.UserID == host {
			t.Errorf("ホストにお知らせが作成されている")
		}
		if n.Type != models.NotificationRoomDismissed || n.Title != "部屋「テスト部屋」が解散されました" {
			t.Errorf("種類/タイトルが誤り: %+v", n)
		}
		if n.ActorUserID == nil || *n.ActorUserID != host {
			t.Errorf("操作者がホストになっていない: %+v", n)
		}
	}
}

func TestNotifyRoomAutoDismissed(t *testing.T) {
	fake := &fakeNotificationRepo{}
	svc := NewNotificationService(&repository.Repository{Notification: fake})

	host := uuid.New()
	guest := uuid.New()
	room := &models.Room{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "放置部屋", HostUserID: host}

	if err := svc.NotifyRoomAutoDismissed(room); err != nil {
		t.Fatalf("NotifyRoomAutoDismissed() error = %v", err)
	}
	if err := svc.NotifyRoomAutoDismissedToMembers(room, []models.RoomMember{{UserID: host, IsHost: true}, {UserID: guest}}); err != nil {
		t.Fatalf("NotifyRoomAutoDismissedToMembers() error = %v", err)
	}

	if len(fake.created) != 2 {
		t.Fatalf("作成されたお知らせ = %d 件, want 2（ホスト向け1 + メンバー向け1）", len(fake.created))
	}
	if fake.created[0].UserID != host || fake.created[0].Title != "部屋「放置部屋」が自動的に削除されました" {
		t.Errorf("ホスト向けの内容が誤り: %+v", fake.created[0])
	}
	if fake.created[1].UserID != guest || fake.created[1].Title != "参加していた部屋「放置部屋」が自動的に削除されました" {
		t.Errorf("メンバー向けの内容が誤り: %+v", fake.created[1])
	}
}
