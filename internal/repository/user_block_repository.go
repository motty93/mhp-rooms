package repository

import (
	"errors"
	"fmt"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userBlockRepository struct {
	db *postgres.DB
}

func NewUserBlockRepository(db *postgres.DB) UserBlockRepository {
	return &userBlockRepository{db: db}
}

// CreateBlock は新しいブロック関係を作成します
func (r *userBlockRepository) CreateBlock(block *models.UserBlock) error {
	if block.BlockerUserID == block.BlockedUserID {
		return errors.New("自分自身をブロックすることはできません")
	}

	// 既に存在するかチェック
	var existing models.UserBlock
	err := r.db.GetConn().Where("blocker_user_id = ? AND blocked_user_id = ?",
		block.BlockerUserID, block.BlockedUserID).First(&existing).Error

	if err == nil {
		return errors.New("既にブロック済みです")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("ブロック関係の確認に失敗しました: %w", err)
	}

	return r.db.GetConn().Create(block).Error
}

// DeleteBlock は指定されたブロック関係を削除します
func (r *userBlockRepository) DeleteBlock(blockerUserID, blockedUserID uuid.UUID) error {
	result := r.db.GetConn().Where("blocker_user_id = ? AND blocked_user_id = ?",
		blockerUserID, blockedUserID).Delete(&models.UserBlock{})

	if result.Error != nil {
		return fmt.Errorf("ブロック関係の削除に失敗しました: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("削除対象のブロック関係が見つかりません")
	}

	return nil
}

// IsBlocked は指定されたユーザーがブロックされているかチェックします
func (r *userBlockRepository) IsBlocked(blockerUserID, blockedUserID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetConn().Model(&models.UserBlock{}).
		Where("blocker_user_id = ? AND blocked_user_id = ?", blockerUserID, blockedUserID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("ブロック状態の確認に失敗しました: %w", err)
	}

	return count > 0, nil
}

// CheckBlockRelationship は2人のユーザー間のブロック関係を双方向で確認します
// 戻り値: (isBlockedByTarget, isBlockingTarget, error)
func (r *userBlockRepository) CheckBlockRelationship(userID, targetUserID uuid.UUID) (bool, bool, error) {
	if userID == targetUserID {
		return false, false, nil
	}

	// userIDがtargetUserIDをブロックしているかチェック
	isBlockingTarget, err := r.IsBlocked(userID, targetUserID)
	if err != nil {
		return false, false, fmt.Errorf("ブロック関係の確認に失敗しました: %w", err)
	}

	// targetUserIDがuserIDをブロックしているかチェック
	isBlockedByTarget, err := r.IsBlocked(targetUserID, userID)
	if err != nil {
		return false, false, fmt.Errorf("ブロック関係の確認に失敗しました: %w", err)
	}

	return isBlockedByTarget, isBlockingTarget, nil
}

// CheckRoomMemberBlocks は指定ユーザーと部屋のメンバーとの間にブロック関係があるかチェックします
func (r *userBlockRepository) CheckRoomMemberBlocks(userID, roomID uuid.UUID) ([]models.User, error) {
	var blockedUsers []models.User

	// ルームメンバーを取得し、ブロック関係をチェック
	query := `
		SELECT DISTINCT u.id, u.display_name, u.supabase_user_id, u.email, u.created_at, u.updated_at
		FROM users u
		INNER JOIN room_members rm ON u.id = rm.user_id
		WHERE rm.room_id = ? AND rm.left_at IS NULL AND u.id != ?
		AND (
			EXISTS (
				SELECT 1 FROM user_blocks ub 
				WHERE ub.blocker_user_id = u.id AND ub.blocked_user_id = ?
			)
			OR EXISTS (
				SELECT 1 FROM user_blocks ub 
				WHERE ub.blocker_user_id = ? AND ub.blocked_user_id = u.id
			)
		)
	`

	err := r.db.GetConn().Raw(query, roomID, userID, userID, userID).Scan(&blockedUsers).Error
	if err != nil {
		return nil, fmt.Errorf("ルームメンバーのブロック関係確認に失敗しました: %w", err)
	}

	return blockedUsers, nil
}

// GetBlockedUsers は指定ユーザーがブロックしているユーザーリストを取得します
func (r *userBlockRepository) GetBlockedUsers(blockerUserID uuid.UUID) ([]models.User, error) {
	var blockedUsers []models.User

	err := r.db.GetConn().Table("users").
		Joins("INNER JOIN user_blocks ON users.id = user_blocks.blocked_user_id").
		Where("user_blocks.blocker_user_id = ?", blockerUserID).
		Select("users.*").
		Find(&blockedUsers).Error

	if err != nil {
		return nil, fmt.Errorf("ブロック済みユーザーの取得に失敗しました: %w", err)
	}

	return blockedUsers, nil
}

// GetBlockingUsers は指定ユーザーをブロックしているユーザーリストを取得します
func (r *userBlockRepository) GetBlockingUsers(blockedUserID uuid.UUID) ([]models.User, error) {
	var blockingUsers []models.User

	err := r.db.GetConn().Table("users").
		Joins("INNER JOIN user_blocks ON users.id = user_blocks.blocker_user_id").
		Where("user_blocks.blocked_user_id = ?", blockedUserID).
		Select("users.*").
		Find(&blockingUsers).Error

	if err != nil {
		return nil, fmt.Errorf("ブロックしているユーザーの取得に失敗しました: %w", err)
	}

	return blockingUsers, nil
}
