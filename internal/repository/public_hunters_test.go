package repository

import (
	"testing"
	"time"

	"mhp-rooms/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type publicHunterTestDB struct{ conn *gorm.DB }

func (d publicHunterTestDB) GetConn() *gorm.DB { return d.conn }
func (d publicHunterTestDB) Close() error      { return nil }
func (d publicHunterTestDB) GetType() string   { return "sqlite" }

func TestPublicHunterQueries(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.UserActivity{}); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(publicHunterTestDB{conn: db})
	now := time.Now().UTC()
	active := newPublicHunterTestUser("アクティブ太郎", "active", true, now.Add(-48*time.Hour))
	rooms := newPublicHunterTestUser("部屋職人", "rooms", true, now.Add(-24*time.Hour))
	inactive := newPublicHunterTestUser("非公開", "hidden", false, now)
	for _, user := range []*models.User{active, rooms, inactive} {
		if err := repo.User.CreateUser(user); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Model(inactive).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	activities := []models.UserActivity{
		{BaseModel: models.BaseModel{CreatedAt: now.Add(-time.Hour)}, UserID: active.ID, ActivityType: models.ActivityRoomJoin, Title: "部屋に参加"},
		{BaseModel: models.BaseModel{CreatedAt: now}, UserID: active.ID, ActivityType: models.ActivityRoomLeave, Title: "退出"},
		{BaseModel: models.BaseModel{CreatedAt: now.Add(-2 * time.Hour)}, UserID: rooms.ID, ActivityType: models.ActivityRoomCreate, Title: "部屋を作成"},
		{BaseModel: models.BaseModel{CreatedAt: now.Add(-3 * time.Hour)}, UserID: rooms.ID, ActivityType: models.ActivityRoomCreate, Title: "もう一部屋"},
		{BaseModel: models.BaseModel{CreatedAt: now.Add(-30 * time.Minute)}, UserID: inactive.ID, ActivityType: models.ActivityRoomCreate, Title: "非公開の部屋"},
	}
	for i := range activities {
		if err := repo.UserActivity.CreateActivity(&activities[i]); err != nil {
			t.Fatal(err)
		}
	}

	for _, tc := range []struct {
		sort string
		want uuid.UUID
	}{{"recent", active.ID}, {"rooms", rooms.ID}, {"joined", rooms.ID}} {
		hunters, err := repo.User.ListPublicHunters(PublicHunterListParams{Sort: tc.sort, Limit: 20})
		if err != nil {
			t.Fatalf("%s: %v", tc.sort, err)
		}
		if len(hunters) != 2 || hunters[0].ID != tc.want {
			t.Fatalf("%s: %#v", tc.sort, hunters)
		}
	}
	matched, err := repo.User.ListPublicHunters(PublicHunterListParams{Query: "太郎", Limit: 20})
	if err != nil || len(matched) != 1 || matched[0].ID != active.ID {
		t.Fatalf("検索結果: %#v, err=%v", matched, err)
	}
	total, err := repo.User.CountPublicHunters("")
	if err != nil || total != 2 {
		t.Fatalf("公開ハンター数 = %d, err=%v", total, err)
	}
}

func newPublicHunterTestUser(name, username string, active bool, createdAt time.Time) *models.User {
	return &models.User{
		BaseModel:      models.BaseModel{ID: uuid.New(), CreatedAt: createdAt, UpdatedAt: createdAt},
		SupabaseUserID: uuid.New(),
		Email:          username + "@example.test",
		Username:       &username,
		DisplayName:    name,
		IsActive:       active,
		Role:           models.RoleUser,
	}
}
