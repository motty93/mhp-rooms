package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/database"
	"mhp-rooms/internal/models"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// User関連メソッド

func (r *Repository) CreateUser(user *models.User) error {
	return r.db.GetConn().Create(user).Error
}

func (r *Repository) FindUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindUserBySupabaseUserID(supabaseUserID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().Where("supabase_user_id = ?", supabaseUserID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(user *models.User) error {
	return r.db.GetConn().Save(user).Error
}

func (r *Repository) GetActiveUsers(limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.GetConn().
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

// GameVersion関連メソッド

func (r *Repository) FindGameVersionByID(id uuid.UUID) (*models.GameVersion, error) {
	var gameVersion models.GameVersion
	err := r.db.GetConn().Where("id = ?", id).First(&gameVersion).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ゲームバージョンが見つかりません")
		}
		return nil, err
	}
	return &gameVersion, nil
}

func (r *Repository) FindGameVersionByCode(code string) (*models.GameVersion, error) {
	var gameVersion models.GameVersion
	err := r.db.GetConn().Where("code = ?", code).First(&gameVersion).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ゲームバージョンが見つかりません")
		}
		return nil, err
	}
	return &gameVersion, nil
}

func (r *Repository) GetActiveGameVersions() ([]models.GameVersion, error) {
	var versions []models.GameVersion
	err := r.db.GetConn().
		Where("is_active = ?", true).
		Order("display_order ASC").
		Find(&versions).Error
	return versions, err
}

// Room関連メソッド

func (r *Repository) CreateRoom(room *models.Room) error {
	return r.db.GetConn().Create(room).Error
}

func (r *Repository) FindRoomByID(id uuid.UUID) (*models.Room, error) {
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

func (r *Repository) FindRoomByRoomCode(roomCode string) (*models.Room, error) {
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

func (r *Repository) GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error) {
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

func (r *Repository) UpdateRoom(room *models.Room) error {
	return r.db.GetConn().Save(room).Error
}

func (r *Repository) UpdateRoomStatus(id uuid.UUID, status string) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) IncrementRoomPlayerCount(id uuid.UUID) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Update("current_players", gorm.Expr("current_players + ?", 1)).Error
}

func (r *Repository) DecrementRoomPlayerCount(id uuid.UUID) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Where("current_players > ?", 0).
		Update("current_players", gorm.Expr("current_players - ?", 1)).Error
}

func (r *Repository) JoinRoom(roomID, userID uuid.UUID, password string) error {
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

func (r *Repository) LeaveRoom(roomID, userID uuid.UUID) error {
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