package repository

import (
	"errors"
	"github.com/google/uuid"
	"mhp-rooms/internal/models"
)

var ErrNotFound = errors.New("not found")

type ReactionRepository interface {
	GetMessageReactions(messageID uuid.UUID, userID *uuid.UUID) ([]models.MessageReactionCount, error)
	AddReaction(reaction *models.MessageReaction) error
	RemoveReaction(messageID, userID uuid.UUID, reactionType string) error
	GetReactionTypes() ([]models.ReactionType, error)
	CheckMessageExists(messageID uuid.UUID) error
	CheckReactionTypeExists(code string) error
}

type reactionRepository struct {
	db DBInterface
}

func NewReactionRepository(db DBInterface) ReactionRepository {
	return &reactionRepository{db: db}
}

func (r *reactionRepository) GetMessageReactions(messageID uuid.UUID, userID *uuid.UUID) ([]models.MessageReactionCount, error) {
	var reactionCounts []models.MessageReactionCount
	query := `
		SELECT 
			mr.message_id,
			mr.reaction_type,
			rt.emoji,
			rt.name AS reaction_name,
			COUNT(mr.user_id) AS reaction_count,
			ARRAY_AGG(mr.user_id ORDER BY mr.created_at) AS user_ids
		FROM message_reactions mr
		JOIN reaction_types rt ON mr.reaction_type = rt.code
		WHERE mr.message_id = ? AND rt.is_active = true
		GROUP BY mr.message_id, mr.reaction_type, rt.emoji, rt.name
		ORDER BY rt.display_order`

	if err := r.db.GetConn().Raw(query, messageID).Scan(&reactionCounts).Error; err != nil {
		return nil, err
	}

	// ユーザーがログインしている場合、そのユーザーのリアクション状態をチェック
	if userID != nil {
		for i := range reactionCounts {
			for _, uid := range reactionCounts[i].UserIDs {
				if uid == *userID {
					reactionCounts[i].HasReacted = true
					break
				}
			}
			// セキュリティのため、他のユーザーIDは削除
			reactionCounts[i].UserIDs = nil
		}
	}

	return reactionCounts, nil
}

func (r *reactionRepository) AddReaction(reaction *models.MessageReaction) error {
	return r.db.GetConn().Create(reaction).Error
}

func (r *reactionRepository) RemoveReaction(messageID, userID uuid.UUID, reactionType string) error {
	result := r.db.GetConn().Where("message_id = ? AND user_id = ? AND reaction_type = ?",
		messageID, userID, reactionType).Delete(&models.MessageReaction{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *reactionRepository) GetReactionTypes() ([]models.ReactionType, error) {
	var reactionTypes []models.ReactionType
	if err := r.db.GetConn().Where("is_active = ?", true).Order("display_order").Find(&reactionTypes).Error; err != nil {
		return nil, err
	}
	return reactionTypes, nil
}

func (r *reactionRepository) CheckMessageExists(messageID uuid.UUID) error {
	var message models.RoomMessage
	return r.db.GetConn().First(&message, messageID).Error
}

func (r *reactionRepository) CheckReactionTypeExists(code string) error {
	var reactionType models.ReactionType
	return r.db.GetConn().Where("code = ? AND is_active = ?", code, true).First(&reactionType).Error
}
