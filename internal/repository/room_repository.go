package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/database"
	"mhp-rooms/internal/models"
)

// roomRepository はルーム関連の操作を行うリポジトリの実装
type roomRepository struct {
	db *database.DB
}

// NewRoomRepository は新しいRoomRepositoryインスタンスを作成
func NewRoomRepository(db *database.DB) RoomRepository {
	return &roomRepository{db: db}
}

// CreateRoom はルームを作成
func (r *roomRepository) CreateRoom(room *models.Room) error {
	return r.db.GetConn().Create(room).Error
}

// FindRoomByID はIDでルームを検索
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

// FindRoomByRoomCode はルームコードでルームを検索
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

// GetActiveRooms はアクティブなルーム一覧を取得
func (r *roomRepository) GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error) {
	var rooms []models.Room
	query := r.db.GetConn().
		Preload("GameVersion").
		Preload("Host").
		Where("is_active = ?", true).
		Where("status IN ?", []string{"waiting", "playing"})

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

// UpdateRoom はルーム情報を更新
func (r *roomRepository) UpdateRoom(room *models.Room) error {
	return r.db.GetConn().Save(room).Error
}

// UpdateRoomStatus はルームのステータスを更新
func (r *roomRepository) UpdateRoomStatus(id uuid.UUID, status string) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// IncrementRoomPlayerCount はルームの参加者数を増やす
func (r *roomRepository) IncrementRoomPlayerCount(id uuid.UUID) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Update("current_players", gorm.Expr("current_players + ?", 1)).Error
}

// DecrementRoomPlayerCount はルームの参加者数を減らす
func (r *roomRepository) DecrementRoomPlayerCount(id uuid.UUID) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Where("current_players > ?", 0).
		Update("current_players", gorm.Expr("current_players - ?", 1)).Error
}

// JoinRoom はユーザーをルームに参加させる
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

// LeaveRoom はユーザーをルームから退出させる
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