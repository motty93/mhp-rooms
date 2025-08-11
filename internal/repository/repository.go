package repository

import (
	"github.com/google/uuid"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

type Repository struct {
	User          UserRepository
	GameVersion   GameVersionRepository
	Platform      PlatformRepository
	Room          RoomRepository
	PasswordReset PasswordResetRepository
	PlayerName    PlayerNameRepository
	Reaction      ReactionRepository
	RoomMessage   RoomMessageRepository
	UserBlock     UserBlockRepository
}

func NewRepository(db *postgres.DB) *Repository {
	return &Repository{
		User:          NewUserRepository(db),
		GameVersion:   NewGameVersionRepository(db),
		Platform:      NewPlatformRepository(db),
		Room:          NewRoomRepository(db),
		PasswordReset: NewPasswordResetRepository(db),
		PlayerName:    NewPlayerNameRepository(db),
		Reaction:      NewReactionRepository(db),
		RoomMessage:   NewRoomMessageRepository(db),
		UserBlock:     NewUserBlockRepository(db),
	}
}

// 後方互換性のための委譲メソッド（段階的に削除予定）

func (r *Repository) CreateUser(user *models.User) error {
	return r.User.CreateUser(user)
}

func (r *Repository) FindUserByID(id uuid.UUID) (*models.User, error) {
	return r.User.FindUserByID(id)
}

func (r *Repository) FindUserBySupabaseUserID(supabaseUserID uuid.UUID) (*models.User, error) {
	return r.User.FindUserBySupabaseUserID(supabaseUserID)
}

func (r *Repository) FindUserByEmail(email string) (*models.User, error) {
	return r.User.FindUserByEmail(email)
}

func (r *Repository) UpdateUser(user *models.User) error {
	return r.User.UpdateUser(user)
}

func (r *Repository) GetActiveUsers(limit, offset int) ([]models.User, error) {
	return r.User.GetActiveUsers(limit, offset)
}

func (r *Repository) FindGameVersionByID(id uuid.UUID) (*models.GameVersion, error) {
	return r.GameVersion.FindGameVersionByID(id)
}

func (r *Repository) FindGameVersionByCode(code string) (*models.GameVersion, error) {
	return r.GameVersion.FindGameVersionByCode(code)
}

func (r *Repository) GetActiveGameVersions() ([]models.GameVersion, error) {
	return r.GameVersion.GetActiveGameVersions()
}

func (r *Repository) GetActivePlatforms() ([]models.Platform, error) {
	return r.Platform.GetActivePlatforms()
}

func (r *Repository) CreateRoom(room *models.Room) error {
	return r.Room.CreateRoom(room)
}

func (r *Repository) FindRoomByID(id uuid.UUID) (*models.Room, error) {
	return r.Room.FindRoomByID(id)
}

func (r *Repository) FindRoomByRoomCode(roomCode string) (*models.Room, error) {
	return r.Room.FindRoomByRoomCode(roomCode)
}

func (r *Repository) GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error) {
	return r.Room.GetActiveRooms(gameVersionID, limit, offset)
}

func (r *Repository) UpdateRoom(room *models.Room) error {
	return r.Room.UpdateRoom(room)
}

func (r *Repository) DismissRoom(id uuid.UUID) error {
	return r.Room.DismissRoom(id)
}

func (r *Repository) ToggleRoomClosed(id uuid.UUID, isClosed bool) error {
	return r.Room.ToggleRoomClosed(id, isClosed)
}

func (r *Repository) IncrementRoomPlayerCount(id uuid.UUID) error {
	return r.Room.IncrementRoomPlayerCount(id)
}

func (r *Repository) DecrementRoomPlayerCount(id uuid.UUID) error {
	return r.Room.DecrementRoomPlayerCount(id)
}

func (r *Repository) JoinRoom(roomID, userID uuid.UUID, password string) error {
	return r.Room.JoinRoom(roomID, userID, password)
}

func (r *Repository) LeaveRoom(roomID, userID uuid.UUID) error {
	return r.Room.LeaveRoom(roomID, userID)
}

// PlayerName関連の委譲メソッド

func (r *Repository) CreatePlayerName(playerName *models.PlayerName) error {
	return r.PlayerName.CreatePlayerName(playerName)
}

func (r *Repository) UpdatePlayerName(playerName *models.PlayerName) error {
	return r.PlayerName.UpdatePlayerName(playerName)
}

func (r *Repository) FindPlayerNameByUserAndGame(userID, gameVersionID uuid.UUID) (*models.PlayerName, error) {
	return r.PlayerName.FindPlayerNameByUserAndGame(userID, gameVersionID)
}

func (r *Repository) FindAllPlayerNamesByUser(userID uuid.UUID) ([]models.PlayerName, error) {
	return r.PlayerName.FindAllPlayerNamesByUser(userID)
}

func (r *Repository) DeletePlayerName(id uuid.UUID) error {
	return r.PlayerName.DeletePlayerName(id)
}

func (r *Repository) UpsertPlayerName(playerName *models.PlayerName) error {
	return r.PlayerName.UpsertPlayerName(playerName)
}
