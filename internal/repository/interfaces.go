package repository

import (
	"github.com/google/uuid"
	"mhp-rooms/internal/models"
)

// UserRepository はユーザー関連の操作を定義するインターフェース
type UserRepository interface {
	CreateUser(user *models.User) error
	FindUserByID(id uuid.UUID) (*models.User, error)
	FindUserBySupabaseUserID(supabaseUserID uuid.UUID) (*models.User, error)
	FindUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	GetActiveUsers(limit, offset int) ([]models.User, error)
}

// GameVersionRepository はゲームバージョン関連の操作を定義するインターフェース
type GameVersionRepository interface {
	FindGameVersionByID(id uuid.UUID) (*models.GameVersion, error)
	FindGameVersionByCode(code string) (*models.GameVersion, error)
	GetActiveGameVersions() ([]models.GameVersion, error)
}

// RoomRepository はルーム関連の操作を定義するインターフェース
type RoomRepository interface {
	CreateRoom(room *models.Room) error
	FindRoomByID(id uuid.UUID) (*models.Room, error)
	FindRoomByRoomCode(roomCode string) (*models.Room, error)
	GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error)
	UpdateRoom(room *models.Room) error
	UpdateRoomStatus(id uuid.UUID, status string) error
	IncrementRoomPlayerCount(id uuid.UUID) error
	DecrementRoomPlayerCount(id uuid.UUID) error
	JoinRoom(roomID, userID uuid.UUID, password string) error
	LeaveRoom(roomID, userID uuid.UUID) error
}