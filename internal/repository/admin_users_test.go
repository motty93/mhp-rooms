package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
)

type adminUsersTestDB struct{ conn *gorm.DB }

func (d adminUsersTestDB) GetConn() *gorm.DB { return d.conn }
func (d adminUsersTestDB) Close() error      { return nil }
func (d adminUsersTestDB) GetType() string   { return "sqlite" }

func TestAdminUserQueries(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Room{}, &models.UserActivity{}, &models.UserReport{}); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(adminUsersTestDB{conn: db})
	now := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	username := "room-master"
	roomHost := newAdminUserTestUser("部屋職人", &username, "rooms@example.test", now.Add(-3*time.Hour))
	reported := newAdminUserTestUser("通報対象", nil, "reports@example.test", now.Add(-2*time.Hour))
	inactive := newAdminUserTestUser("休止中", adminUserTestStringPtr("inactive-hunter"), "inactive@example.test", now.Add(-time.Hour))
	newest := newAdminUserTestUser("新規ユーザー", adminUserTestStringPtr("newest"), "newest@example.test", now)
	for _, user := range []*models.User{roomHost, reported, inactive, newest} {
		if err := repo.User.CreateUser(user); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Model(inactive).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		room := &models.Room{
			BaseModel:     models.BaseModel{ID: uuid.New(), CreatedAt: now.Add(time.Duration(i) * time.Minute)},
			RoomCode:      "ADMIN-ROOM-" + string(rune('A'+i)),
			Name:          "管理用部屋",
			GameVersionID: uuid.New(),
			HostUserID:    roomHost.ID,
			MaxPlayers:    4,
			IsActive:      false,
		}
		if err := db.Create(room).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		report := &models.UserReport{
			BaseModel:      models.BaseModel{ID: uuid.New(), CreatedAt: now.Add(time.Duration(i) * time.Minute)},
			ReporterUserID: roomHost.ID,
			ReportedUserID: reported.ID,
			Reason:         models.ReasonSpam,
			Description:    "テスト通報",
			Status:         models.ReportStatusPending,
		}
		if err := db.Create(report).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, activityAt := range []time.Time{now.Add(-4 * time.Hour), now.Add(-30 * time.Minute)} {
		activity := &models.UserActivity{BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: activityAt}, UserID: roomHost.ID, ActivityType: models.ActivityRoomCreate, Title: "活動"}
		if err := db.Create(activity).Error; err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name   string
		params AdminUserListParams
		want   uuid.UUID
	}{
		{"表示名検索", AdminUserListParams{Query: "休止", Limit: 20}, inactive.ID},
		{"ユーザー名検索", AdminUserListParams{Query: "ROOM-MASTER", Limit: 20}, roomHost.ID},
		{"メール検索", AdminUserListParams{Query: "REPORTS@EXAMPLE", Limit: 20}, reported.ID},
		{"部屋作成数順", AdminUserListParams{Sort: "rooms", Limit: 20}, roomHost.ID},
		{"通報数順", AdminUserListParams{Sort: "reports", Limit: 20}, reported.ID},
		{"登録日順", AdminUserListParams{Sort: "joined", Limit: 20}, newest.ID},
		{"不正な並び順は登録日順", AdminUserListParams{Sort: "unknown", Limit: 20}, newest.ID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := repo.User.ListAdminUsers(tt.params)
			if err != nil {
				t.Fatal(err)
			}
			if len(users) == 0 || users[0].ID != tt.want {
				t.Fatalf("先頭ユーザー = %#v, want %s", users, tt.want)
			}
		})
	}

	users, err := repo.User.ListAdminUsers(AdminUserListParams{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 4 {
		t.Fatalf("全ユーザー数 = %d, want 4", len(users))
	}
	var roomHostRow, inactiveRow *AdminUser
	for i := range users {
		if users[i].ID == roomHost.ID {
			roomHostRow = &users[i]
		}
		if users[i].ID == inactive.ID {
			inactiveRow = &users[i]
		}
	}
	if roomHostRow == nil || roomHostRow.RoomCount != 2 || roomHostRow.LastActivityAt == nil || !roomHostRow.LastActivityAt.Equal(now.Add(-30*time.Minute)) {
		t.Fatalf("部屋ホストの集計 = %#v", roomHostRow)
	}
	if inactiveRow == nil || inactiveRow.Username == nil || *inactiveRow.Username != "inactive-hunter" || inactiveRow.RoomCount != 0 || inactiveRow.ReportCount != 0 || inactiveRow.LastActivityAt != nil {
		t.Fatalf("無効ユーザーまたはゼロ集計 = %#v", inactiveRow)
	}
	count, err := repo.User.CountAdminUsers("")
	if err != nil || count != 4 {
		t.Fatalf("全ユーザー件数 = %d, err=%v", count, err)
	}
	count, err = repo.User.CountAdminUsers("rooms@example")
	if err != nil || count != 1 {
		t.Fatalf("メール検索件数 = %d, err=%v", count, err)
	}
}

func TestAdminRoomHostFilter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.GameVersion{}, &models.Room{}); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(adminUsersTestDB{conn: db})
	now := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	hostA := newAdminUserTestUser("ホストA", adminUserTestStringPtr("host-a"), "host-a@example.test", now)
	hostB := newAdminUserTestUser("ホストB", adminUserTestStringPtr("host-b"), "host-b@example.test", now)
	for _, user := range []*models.User{hostA, hostB} {
		if err := repo.User.CreateUser(user); err != nil {
			t.Fatal(err)
		}
	}

	for _, room := range []*models.Room{
		{BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: now}, RoomCode: "HOST-A-1", Name: "ホストAの部屋1", GameVersionID: uuid.New(), HostUserID: hostA.ID, MaxPlayers: 4, IsActive: true},
		{BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: now.Add(-time.Minute)}, RoomCode: "HOST-A-2", Name: "ホストAの部屋2", GameVersionID: uuid.New(), HostUserID: hostA.ID, MaxPlayers: 4, IsActive: false},
		{BaseModel: models.BaseModel{ID: uuid.New(), CreatedAt: now.Add(-2 * time.Minute)}, RoomCode: "HOST-B-1", Name: "ホストBの部屋", GameVersionID: uuid.New(), HostUserID: hostB.ID, MaxPlayers: 4, IsActive: true},
	} {
		if err := db.Create(room).Error; err != nil {
			t.Fatal(err)
		}
	}

	rooms, err := repo.Room.GetAllRoomsForAdmin(20, 0, &hostA.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 2 {
		t.Fatalf("ホストAの部屋数 = %d, want 2", len(rooms))
	}
	for _, room := range rooms {
		if room.HostUserID != hostA.ID {
			t.Errorf("別ホストの部屋を取得: %+v", room)
		}
	}

	count, err := repo.Room.CountAllRooms(&hostA.ID)
	if err != nil || count != 2 {
		t.Fatalf("ホストAの部屋件数 = %d, err=%v, want 2", count, err)
	}
	count, err = repo.Room.CountAllRooms(nil)
	if err != nil || count != 3 {
		t.Fatalf("全部屋件数 = %d, err=%v, want 3", count, err)
	}
}

func newAdminUserTestUser(displayName string, username *string, email string, createdAt time.Time) *models.User {
	return &models.User{
		BaseModel:      models.BaseModel{ID: uuid.New(), CreatedAt: createdAt, UpdatedAt: createdAt},
		SupabaseUserID: uuid.New(),
		Email:          email,
		Username:       username,
		DisplayName:    displayName,
		IsActive:       true,
		Role:           models.RoleUser,
	}
}

func adminUserTestStringPtr(value string) *string {
	return &value
}
