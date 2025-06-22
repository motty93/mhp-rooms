package repository

import (
	"github.com/google/uuid"
	"mhp-rooms/internal/database"
	"mhp-rooms/internal/models"
)

// Repository は各テーブルのリポジトリを統合する構造体
type Repository struct {
	User        UserRepository
	GameVersion GameVersionRepository
	Room        RoomRepository
}

// NewRepository は新しいRepositoryインスタンスを作成
func NewRepository(db *database.DB) *Repository {
	return &Repository{
		User:        NewUserRepository(db),
		GameVersion: NewGameVersionRepository(db),
		Room:        NewRoomRepository(db),
	}
}

// 後方互換性のための委譲メソッド（段階的に削除予定）

// CreateUser はUser.CreateUserに委譲
func (r *Repository) CreateUser(user *models.User) error {
	return r.User.CreateUser(user)
}

// FindUserByID はUser.FindUserByIDに委譲
func (r *Repository) FindUserByID(id uuid.UUID) (*models.User, error) {
	return r.User.FindUserByID(id)
}

// FindUserBySupabaseUserID はUser.FindUserBySupabaseUserIDに委譲
func (r *Repository) FindUserBySupabaseUserID(supabaseUserID uuid.UUID) (*models.User, error) {
	return r.User.FindUserBySupabaseUserID(supabaseUserID)
}

// FindUserByEmail はUser.FindUserByEmailに委譲
func (r *Repository) FindUserByEmail(email string) (*models.User, error) {
	return r.User.FindUserByEmail(email)
}

// UpdateUser はUser.UpdateUserに委譲
func (r *Repository) UpdateUser(user *models.User) error {
	return r.User.UpdateUser(user)
}

// GetActiveUsers はUser.GetActiveUsersに委譲
func (r *Repository) GetActiveUsers(limit, offset int) ([]models.User, error) {
	return r.User.GetActiveUsers(limit, offset)
}

// FindGameVersionByID はGameVersion.FindGameVersionByIDに委譲
func (r *Repository) FindGameVersionByID(id uuid.UUID) (*models.GameVersion, error) {
	return r.GameVersion.FindGameVersionByID(id)
}

// FindGameVersionByCode はGameVersion.FindGameVersionByCodeに委譲
func (r *Repository) FindGameVersionByCode(code string) (*models.GameVersion, error) {
	return r.GameVersion.FindGameVersionByCode(code)
}

// GetActiveGameVersions はGameVersion.GetActiveGameVersionsに委譲
func (r *Repository) GetActiveGameVersions() ([]models.GameVersion, error) {
	return r.GameVersion.GetActiveGameVersions()
}

// CreateRoom はRoom.CreateRoomに委譲
func (r *Repository) CreateRoom(room *models.Room) error {
	return r.Room.CreateRoom(room)
}

// FindRoomByID はRoom.FindRoomByIDに委譲
func (r *Repository) FindRoomByID(id uuid.UUID) (*models.Room, error) {
	return r.Room.FindRoomByID(id)
}

// FindRoomByRoomCode はRoom.FindRoomByRoomCodeに委譲
func (r *Repository) FindRoomByRoomCode(roomCode string) (*models.Room, error) {
	return r.Room.FindRoomByRoomCode(roomCode)
}

// GetActiveRooms はRoom.GetActiveRoomsに委譲
func (r *Repository) GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error) {
	return r.Room.GetActiveRooms(gameVersionID, limit, offset)
}

// UpdateRoom はRoom.UpdateRoomに委譲
func (r *Repository) UpdateRoom(room *models.Room) error {
	return r.Room.UpdateRoom(room)
}

// UpdateRoomStatus はRoom.UpdateRoomStatusに委譲
func (r *Repository) UpdateRoomStatus(id uuid.UUID, status string) error {
	return r.Room.UpdateRoomStatus(id, status)
}

// IncrementRoomPlayerCount はRoom.IncrementRoomPlayerCountに委譲
func (r *Repository) IncrementRoomPlayerCount(id uuid.UUID) error {
	return r.Room.IncrementRoomPlayerCount(id)
}

// DecrementRoomPlayerCount はRoom.DecrementRoomPlayerCountに委譲
func (r *Repository) DecrementRoomPlayerCount(id uuid.UUID) error {
	return r.Room.DecrementRoomPlayerCount(id)
}

// JoinRoom はRoom.JoinRoomに委譲
func (r *Repository) JoinRoom(roomID, userID uuid.UUID, password string) error {
	return r.Room.JoinRoom(roomID, userID, password)
}

// LeaveRoom はRoom.LeaveRoomに委譲
func (r *Repository) LeaveRoom(roomID, userID uuid.UUID) error {
	return r.Room.LeaveRoom(roomID, userID)
}
