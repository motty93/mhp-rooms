package repository

import (
	"testing"
	"time"

	"mhp-rooms/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type recentActivityTestDB struct{ conn *gorm.DB }

func (d recentActivityTestDB) GetConn() *gorm.DB { return d.conn }
func (d recentActivityTestDB) Close() error      { return nil }
func (d recentActivityTestDB) GetType() string   { return "sqlite" }

func TestRecentPublicActivityQueries(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.UserActivity{}); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(recentActivityTestDB{conn: db})
	now := time.Now().UTC()
	active := newRecentActivityTestUser("アクティブ太郎", "active", true, now.Add(-48*time.Hour))
	inactive := newRecentActivityTestUser("非公開", "hidden", false, now)
	for _, user := range []*models.User{active, inactive} {
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
		{BaseModel: models.BaseModel{CreatedAt: now.Add(-2 * time.Hour)}, UserID: active.ID, ActivityType: models.ActivityRoomCreate, Title: "部屋を作成"},
		{BaseModel: models.BaseModel{CreatedAt: now.Add(-30 * time.Minute)}, UserID: inactive.ID, ActivityType: models.ActivityRoomCreate, Title: "非公開の部屋"},
	}
	for i := range activities {
		if err := repo.UserActivity.CreateActivity(&activities[i]); err != nil {
			t.Fatal(err)
		}
	}

	recent, err := repo.UserActivity.GetRecentPublicActivities(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 2 {
		t.Fatalf("公開活動 = %d, want 2", len(recent))
	}
	for _, activity := range recent {
		if !models.IsPublicFeedActivity(activity.ActivityType) || activity.User.ID == inactive.ID {
			t.Fatalf("非公開活動または非アクティブユーザーを取得: %#v", activity)
		}
	}
	if _, err := repo.UserActivity.CountPublicActivitiesByTypeSince(models.ActivityRoomLeave, now.Add(-24*time.Hour)); err == nil {
		t.Fatal("非公開活動の集計でエラーが返りませんでした")
	}
	count, err := repo.UserActivity.CountPublicActivitiesByTypeSince(models.ActivityRoomCreate, now.Add(-24*time.Hour))
	if err != nil || count != 1 {
		t.Fatalf("部屋作成数 = %d, err=%v", count, err)
	}
}

func newRecentActivityTestUser(name, username string, active bool, createdAt time.Time) *models.User {
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
