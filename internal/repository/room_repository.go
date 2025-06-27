package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

type roomRepository struct {
	db *postgres.DB
}

func NewRoomRepository(db *postgres.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) CreateRoom(room *models.Room) error {
	return r.db.GetConn().Create(room).Error
}

func (r *roomRepository) FindRoomByID(id uuid.UUID) (*models.Room, error) {
	var room models.Room
	err := r.db.GetConn().
		Preload("GameVersion").
		Preload("Host").
		Where("id = ?", id).
		First(&room).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ルームが見つかりません")
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) FindRoomByRoomCode(roomCode string) (*models.Room, error) {
	var room models.Room
	err := r.db.GetConn().
		Preload("GameVersion").
		Preload("Host").
		Where("room_code = ?", roomCode).
		First(&room).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ルームが見つかりません")
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error) {
	var rooms []models.Room
	query := r.db.GetConn().
		Preload("GameVersion").
		Preload("Host").
		Where("is_active = ?", true)

	if gameVersionID != nil {
		query = query.Where("game_version_id = ?", *gameVersionID)
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rooms).Error
	return rooms, err
}

func (r *roomRepository) UpdateRoom(room *models.Room) error {
	return r.db.GetConn().Save(room).Error
}

func (r *roomRepository) ToggleRoomClosed(id uuid.UUID, isClosed bool) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Update("is_closed", isClosed).Error
}

func (r *roomRepository) IncrementRoomPlayerCount(id uuid.UUID) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Update("current_players", gorm.Expr("current_players + ?", 1)).Error
}

func (r *roomRepository) DecrementRoomPlayerCount(id uuid.UUID) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Where("current_players > ?", 0).
		Update("current_players", gorm.Expr("current_players - ?", 1)).Error
}

func (r *roomRepository) JoinRoom(roomID, userID uuid.UUID, password string) error {
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		var room models.Room
		if err := tx.Where("id = ?", roomID).First(&room).Error; err != nil {
			return err
		}

		if !room.CanJoin() {
			return fmt.Errorf("ルームに参加できません")
		}

		if !room.CheckPassword(password) {
			return fmt.Errorf("パスワードが間違っています")
		}

		var existingMember models.RoomMember
		err := tx.Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "active").
			First(&existingMember).Error
		if err == nil {
			return fmt.Errorf("既にルームに参加しています")
		}

		member := models.RoomMember{
			RoomID: roomID,
			UserID: userID,
			Status: "active",
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		return tx.Model(&models.Room{}).
			Where("id = ?", roomID).
			Update("current_players", gorm.Expr("current_players + ?", 1)).Error
	})
}

func (r *roomRepository) LeaveRoom(roomID, userID uuid.UUID) error {
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.RoomMember{}).
			Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "active").
			Update("status", "left")

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("ルームメンバーが見つかりません")
		}

		return tx.Model(&models.Room{}).
			Where("id = ?", roomID).
			Where("current_players > ?", 0).
			Update("current_players", gorm.Expr("current_players - ?", 1)).Error
	})
}
