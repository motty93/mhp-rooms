package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
)

type roomMessageRepository struct {
	db DBInterface
}

func NewRoomMessageRepository(db DBInterface) RoomMessageRepository {
	return &roomMessageRepository{db: db}
}

func (r *roomMessageRepository) CreateMessage(message *models.RoomMessage) error {
	return r.db.GetConn().Create(message).Error
}

func (r *roomMessageRepository) GetMessages(roomID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.RoomMessage, error) {
	var messages []models.RoomMessage
	query := r.db.GetConn().
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "is_active", "role", "created_at", "updated_at")
		}).
		Where("room_id = ? AND is_deleted = ?", roomID, false).
		Order("created_at DESC").
		Limit(limit)

	if beforeID != nil {
		var beforeMessage models.RoomMessage
		if err := r.db.GetConn().Where("id = ?", *beforeID).First(&beforeMessage).Error; err == nil {
			query = query.Where("created_at < ?", beforeMessage.CreatedAt)
		}
	}

	err := query.Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 時系列順に並べ替え（新しい順から古い順で取得したため）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *roomMessageRepository) DeleteMessage(id uuid.UUID) error {
	return r.db.GetConn().Model(&models.RoomMessage{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}
