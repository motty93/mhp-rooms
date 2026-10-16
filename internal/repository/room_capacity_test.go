package repository

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
)

type roomCapacityTestDB struct{ conn *gorm.DB }

func (d roomCapacityTestDB) GetConn() *gorm.DB { return d.conn }
func (d roomCapacityTestDB) Close() error      { return nil }
func (d roomCapacityTestDB) GetType() string   { return "sqlite" }

func TestRoomCapacity(t *testing.T) {
	t.Run("部屋作成直後の人数はホストを含む 1", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		host := newRoomCapacityTestUser(t, repo, "host")
		room := newRoomCapacityTestRoom(t, repo, host, 2)

		requireRoomCapacityPlayers(t, db, room.ID, 1)
	})

	t.Run("定員に達した部屋には参加できない", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		host := newRoomCapacityTestUser(t, repo, "host")
		guestA := newRoomCapacityTestUser(t, repo, "guest-a")
		guestB := newRoomCapacityTestUser(t, repo, "guest-b")
		room := newRoomCapacityTestRoom(t, repo, host, 2)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestA)
		requireRoomCapacityPlayers(t, db, room.ID, 2)

		requireRoomCapacityError(t, repo.Room.JoinRoom(room.ID, guestB.ID, "", false), "ROOM_FULL:")
		requireRoomCapacityPlayers(t, db, room.ID, 2)
	})

	t.Run("退出して空いた枠に別のユーザーが参加できる", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		host := newRoomCapacityTestUser(t, repo, "host")
		guestA := newRoomCapacityTestUser(t, repo, "guest-a")
		guestB := newRoomCapacityTestUser(t, repo, "guest-b")
		room := newRoomCapacityTestRoom(t, repo, host, 2)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestA)
		requireRoomCapacityPlayers(t, db, room.ID, 2)

		if err := repo.Room.LeaveRoom(room.ID, guestA.ID); err != nil {
			t.Fatal(err)
		}
		requireRoomCapacityPlayers(t, db, room.ID, 1)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestB)
		requireRoomCapacityPlayers(t, db, room.ID, 2)

		requireRoomCapacityError(t, repo.Room.JoinRoom(room.ID, guestA.ID, "", false), "ROOM_FULL:")
		requireRoomCapacityPlayers(t, db, room.ID, 2)

		if err := repo.Room.LeaveRoom(room.ID, guestB.ID); err != nil {
			t.Fatal(err)
		}
		requireRoomCapacityPlayers(t, db, room.ID, 1)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestA)
		requireRoomCapacityPlayers(t, db, room.ID, 2)
	})

	t.Run("キックで空いた枠に別のユーザーが参加できる", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		host := newRoomCapacityTestUser(t, repo, "host")
		guestA := newRoomCapacityTestUser(t, repo, "guest-a")
		guestB := newRoomCapacityTestUser(t, repo, "guest-b")
		room := newRoomCapacityTestRoom(t, repo, host, 2)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestA)
		if err := repo.Room.KickMember(room.ID, guestA.ID); err != nil {
			t.Fatal(err)
		}
		requireRoomCapacityPlayers(t, db, room.ID, 1)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestB)
		requireRoomCapacityPlayers(t, db, room.ID, 2)

		requireRoomCapacityError(t, repo.Room.JoinRoom(room.ID, guestA.ID, "", false), "KICKED:")
		requireRoomCapacityPlayers(t, db, room.ID, 2)
	})

	t.Run("満員でも参加済みのメンバーには ALREADY_JOINED を返す", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		host := newRoomCapacityTestUser(t, repo, "host")
		guestA := newRoomCapacityTestUser(t, repo, "guest-a")
		room := newRoomCapacityTestRoom(t, repo, host, 2)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestA)

		requireRoomCapacityError(t, repo.Room.JoinRoom(room.ID, guestA.ID, "", false), "ALREADY_JOINED:")
		requireRoomCapacityError(t, repo.Room.JoinRoom(room.ID, host.ID, "", false), "ALREADY_JOINED:")
		requireRoomCapacityPlayers(t, db, room.ID, 2)
	})

	t.Run("current_players が古くても満員を判定し、退出で正しい値に直る", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		host := newRoomCapacityTestUser(t, repo, "host")
		guestA := newRoomCapacityTestUser(t, repo, "guest-a")
		guestB := newRoomCapacityTestUser(t, repo, "guest-b")
		room := newRoomCapacityTestRoom(t, repo, host, 2)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestA)
		// ホストを数えていなかった頃のデータを再現する（実際は 2 人なのに列は 1）
		if err := db.Model(&models.Room{}).Where("id = ?", room.ID).Update("current_players", 1).Error; err != nil {
			t.Fatal(err)
		}

		requireRoomCapacityError(t, repo.Room.JoinRoom(room.ID, guestB.ID, "", false), "ROOM_FULL:")
		if got := countRoomCapacityActiveMembers(t, db, room.ID); got != 2 {
			t.Fatalf("拒否された参加でメンバーが増えた: アクティブメンバー数 = %d, want 2", got)
		}

		if err := repo.Room.LeaveRoom(room.ID, guestA.ID); err != nil {
			t.Fatal(err)
		}
		requireRoomCapacityPlayers(t, db, room.ID, 1)

		joinRoomCapacityTestRoom(t, repo, room.ID, guestB)
		requireRoomCapacityPlayers(t, db, room.ID, 2)
	})

	t.Run("他の部屋に参加中のユーザーが満員の部屋へ強制参加しても元の部屋に残る", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		hostX := newRoomCapacityTestUser(t, repo, "host-x")
		hostY := newRoomCapacityTestUser(t, repo, "host-y")
		guest := newRoomCapacityTestUser(t, repo, "guest")
		other := newRoomCapacityTestUser(t, repo, "other")
		roomX := newRoomCapacityTestRoom(t, repo, hostX, 4)
		roomY := newRoomCapacityTestRoom(t, repo, hostY, 2)

		joinRoomCapacityTestRoom(t, repo, roomX.ID, guest)
		joinRoomCapacityTestRoom(t, repo, roomY.ID, other)

		requireRoomCapacityError(t, repo.Room.JoinRoom(roomY.ID, guest.ID, "", true), "ROOM_FULL:")
		requireRoomCapacityError(t, repo.Room.JoinRoom(roomY.ID, guest.ID, "", false), "ROOM_FULL:")

		requireRoomCapacityMemberStatus(t, db, roomX.ID, guest.ID, models.MemberStatusActive)
		requireRoomCapacityPlayers(t, db, roomX.ID, 2)
		requireRoomCapacityPlayers(t, db, roomY.ID, 2)
	})

	t.Run("他の部屋に参加中のユーザーがパスワードを間違えて強制参加しても元の部屋に残る", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		hostX := newRoomCapacityTestUser(t, repo, "host-x")
		hostY := newRoomCapacityTestUser(t, repo, "host-y")
		guest := newRoomCapacityTestUser(t, repo, "guest")
		roomX := newRoomCapacityTestRoom(t, repo, hostX, 4)
		roomY := newRoomCapacityTestRoom(t, repo, hostY, 4)
		setRoomCapacityTestPassword(t, repo, roomY, "secret")

		joinRoomCapacityTestRoom(t, repo, roomX.ID, guest)

		requireRoomCapacityError(t, repo.Room.JoinRoom(roomY.ID, guest.ID, "wrong", true), "パスワードが間違っています")
		requireRoomCapacityError(t, repo.Room.JoinRoom(roomY.ID, guest.ID, "wrong", false), "パスワードが間違っています")

		requireRoomCapacityMemberStatus(t, db, roomX.ID, guest.ID, models.MemberStatusActive)
		requireRoomCapacityPlayers(t, db, roomX.ID, 2)
		requireRoomCapacityPlayers(t, db, roomY.ID, 1)

		if err := repo.Room.JoinRoom(roomY.ID, guest.ID, "secret", true); err != nil {
			t.Fatal(err)
		}
		requireRoomCapacityMemberStatus(t, db, roomX.ID, guest.ID, models.MemberStatusLeft)
		requireRoomCapacityPlayers(t, db, roomY.ID, 2)
	})

	t.Run("強制参加に成功すると元の部屋から退出して移動先に参加する", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		hostX := newRoomCapacityTestUser(t, repo, "host-x")
		hostY := newRoomCapacityTestUser(t, repo, "host-y")
		guest := newRoomCapacityTestUser(t, repo, "guest")
		member := newRoomCapacityTestUser(t, repo, "member")
		roomX := newRoomCapacityTestRoom(t, repo, hostX, 4)
		roomY := newRoomCapacityTestRoom(t, repo, hostY, 4)

		joinRoomCapacityTestRoom(t, repo, roomX.ID, guest)
		joinRoomCapacityTestRoom(t, repo, roomX.ID, member)
		requireRoomCapacityPlayers(t, db, roomX.ID, 3)

		if err := repo.Room.JoinRoom(roomY.ID, guest.ID, "", true); err != nil {
			t.Fatal(err)
		}

		requireRoomCapacityMemberStatus(t, db, roomX.ID, guest.ID, models.MemberStatusLeft)
		requireRoomCapacityPlayers(t, db, roomX.ID, 2)
		requireRoomCapacityMemberStatus(t, db, roomY.ID, guest.ID, models.MemberStatusActive)
		requireRoomCapacityPlayers(t, db, roomY.ID, 2)
		if got := countRoomCapacityLogs(t, db, roomX.ID, guest.ID, "leave"); got != 1 {
			t.Fatalf("部屋 X の leave ログ = %d, want 1", got)
		}
	})

	t.Run("強制参加でなければ他の部屋に参加中のユーザーは参加できない", func(t *testing.T) {
		repo, db := newRoomCapacityTestRepo(t)
		hostX := newRoomCapacityTestUser(t, repo, "host-x")
		hostY := newRoomCapacityTestUser(t, repo, "host-y")
		guest := newRoomCapacityTestUser(t, repo, "guest")
		roomX := newRoomCapacityTestRoom(t, repo, hostX, 4)
		roomY := newRoomCapacityTestRoom(t, repo, hostY, 4)

		joinRoomCapacityTestRoom(t, repo, roomX.ID, guest)

		requireRoomCapacityError(t, repo.Room.JoinRoom(roomY.ID, guest.ID, "", false), "OTHER_ROOM_ACTIVE:")

		requireRoomCapacityMemberStatus(t, db, roomX.ID, guest.ID, models.MemberStatusActive)
		requireRoomCapacityPlayers(t, db, roomX.ID, 2)
		requireRoomCapacityPlayers(t, db, roomY.ID, 1)
	})
}

func newRoomCapacityTestRepo(t *testing.T) (*Repository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.GameVersion{}, &models.Room{}, &models.RoomMember{}, &models.RoomLog{}); err != nil {
		t.Fatal(err)
	}
	return NewRepository(roomCapacityTestDB{conn: db}), db
}

func newRoomCapacityTestUser(t *testing.T, repo *Repository, name string) *models.User {
	t.Helper()
	user := &models.User{
		SupabaseUserID: uuid.New(),
		Email:          name + "@example.test",
		Username:       &name,
		DisplayName:    name,
		IsActive:       true,
		Role:           models.RoleUser,
	}
	if err := repo.User.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	return user
}

func newRoomCapacityTestRoom(t *testing.T, repo *Repository, host *models.User, maxPlayers int) *models.Room {
	t.Helper()
	room := &models.Room{
		RoomCode:      "CAP-" + uuid.NewString()[:8],
		Name:          "定員テスト部屋",
		GameVersionID: uuid.New(),
		HostUserID:    host.ID,
		MaxPlayers:    maxPlayers,
		IsActive:      true,
	}
	if err := repo.Room.CreateRoom(room); err != nil {
		t.Fatal(err)
	}
	return room
}

func setRoomCapacityTestPassword(t *testing.T, repo *Repository, room *models.Room, password string) {
	t.Helper()
	if err := room.SetPassword(password); err != nil {
		t.Fatal(err)
	}
	if err := repo.Room.UpdateRoom(room); err != nil {
		t.Fatal(err)
	}
}

func joinRoomCapacityTestRoom(t *testing.T, repo *Repository, roomID uuid.UUID, user *models.User) {
	t.Helper()
	if err := repo.Room.JoinRoom(roomID, user.ID, "", false); err != nil {
		t.Fatalf("%s の参加に失敗: %v", user.DisplayName, err)
	}
}

func countRoomCapacityActiveMembers(t *testing.T, db *gorm.DB, roomID uuid.UUID) int {
	t.Helper()
	var count int64
	if err := db.Model(&models.RoomMember{}).
		Where("room_id = ? AND status = ?", roomID, models.MemberStatusActive).
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	return int(count)
}

func countRoomCapacityLogs(t *testing.T, db *gorm.DB, roomID, userID uuid.UUID, action string) int {
	t.Helper()
	var count int64
	if err := db.Model(&models.RoomLog{}).
		Where("room_id = ? AND user_id = ? AND action = ?", roomID, userID, action).
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	return int(count)
}

func requireRoomCapacityPlayers(t *testing.T, db *gorm.DB, roomID uuid.UUID, want int) {
	t.Helper()
	var room models.Room
	if err := db.First(&room, "id = ?", roomID).Error; err != nil {
		t.Fatal(err)
	}
	active := countRoomCapacityActiveMembers(t, db, roomID)
	if room.CurrentPlayers != want || active != want {
		t.Fatalf("current_players = %d, アクティブメンバー数 = %d, want %d", room.CurrentPlayers, active, want)
	}
}

func requireRoomCapacityMemberStatus(t *testing.T, db *gorm.DB, roomID, userID uuid.UUID, want string) {
	t.Helper()
	var member models.RoomMember
	if err := db.Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error; err != nil {
		t.Fatalf("メンバーを取得できない: %v", err)
	}
	if member.Status != want {
		t.Fatalf("status = %q, want %q", member.Status, want)
	}
}

func requireRoomCapacityError(t *testing.T, err error, prefix string) {
	t.Helper()
	if err == nil || !strings.HasPrefix(err.Error(), prefix) {
		t.Fatalf("err = %v, want prefix %q", err, prefix)
	}
}
