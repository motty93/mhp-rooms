package repository

import (
	"fmt"
	"time"

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
		// ユーザーの存在確認（開発環境では自動作成）
		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 開発環境でのみダミーユーザーを作成
				user = models.User{
					ID:             userID,
					SupabaseUserID: userID, // 開発用
					Email:          fmt.Sprintf("dev-user-%s@example.com", userID.String()[:8]),
					DisplayName:    fmt.Sprintf("開発ユーザー_%s", userID.String()[:8]),
					IsActive:       true,
					Role:           "user",
				}
				if createErr := tx.Create(&user).Error; createErr != nil {
					return fmt.Errorf("開発用ユーザー作成に失敗しました: %w", createErr)
				}
			} else {
				return err
			}
		}

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

		// アクティブなメンバーかチェック
		var activeMember models.RoomMember
		err := tx.Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "active").
			First(&activeMember).Error
		if err == nil {
			// 既に参加している場合は特別なエラーコードを返す
			return fmt.Errorf("ALREADY_JOINED:既にルームに参加しています")
		}

		// 既存の退室済みメンバーがいるかチェック
		var leftMember models.RoomMember
		err = tx.Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "left").
			First(&leftMember).Error

		if err == nil {
			// 既存の退室済みレコードを再アクティブ化
			if err := tx.Model(&leftMember).Updates(map[string]interface{}{
				"status":    "active",
				"joined_at": time.Now(),
				"left_at":   nil,
			}).Error; err != nil {
				return err
			}
		} else {
			// 新規メンバー作成
			var maxPlayerNumber int
			tx.Model(&models.RoomMember{}).
				Where("room_id = ? AND status = ?", roomID, "active").
				Select("COALESCE(MAX(player_number), 0)").
				Scan(&maxPlayerNumber)
			
			member := models.RoomMember{
				RoomID:       roomID,
				UserID:       userID,
				PlayerNumber: maxPlayerNumber + 1,
				Status:       "active",
				JoinedAt:     time.Now(),
			}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}

		return tx.Model(&models.Room{}).
			Where("id = ?", roomID).
			Update("current_players", gorm.Expr("current_players + ?", 1)).Error
	})
}

func (r *roomRepository) LeaveRoom(roomID, userID uuid.UUID) error {
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&models.RoomMember{}).
			Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "active").
			Updates(map[string]interface{}{
				"status":  "left",
				"left_at": &now,
			})

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

func (r *roomRepository) IsUserJoinedRoom(roomID, userID uuid.UUID) bool {
	var member models.RoomMember
	err := r.db.GetConn().Where("room_id = ? AND user_id = ? AND status = ?", 
		roomID, userID, "active").First(&member).Error
	return err == nil
}
