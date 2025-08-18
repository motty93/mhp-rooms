package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

// userFollowRepository はユーザーフォロー関連の操作を行うリポジトリの実装
type userFollowRepository struct {
	db *postgres.DB
}

// NewUserFollowRepository は新しいUserFollowRepositoryインスタンスを作成
func NewUserFollowRepository(db *postgres.DB) UserFollowRepository {
	return &userFollowRepository{db: db}
}

// CreateFollow フォロー関係を作成
func (r *userFollowRepository) CreateFollow(follow *models.UserFollow) error {
	return r.db.GetConn().Create(follow).Error
}

// DeleteFollow フォロー関係を削除
func (r *userFollowRepository) DeleteFollow(followerUserID, followingUserID uuid.UUID) error {
	result := r.db.GetConn().
		Where("follower_user_id = ? AND following_user_id = ?", followerUserID, followingUserID).
		Delete(&models.UserFollow{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("フォロー関係が見つかりません")
	}
	return nil
}

// GetFollow 特定のフォロー関係を取得
func (r *userFollowRepository) GetFollow(followerUserID, followingUserID uuid.UUID) (*models.UserFollow, error) {
	var follow models.UserFollow
	err := r.db.GetConn().
		Where("follower_user_id = ? AND following_user_id = ?", followerUserID, followingUserID).
		First(&follow).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &follow, nil
}

// UpdateFollowStatus フォローステータスを更新
func (r *userFollowRepository) UpdateFollowStatus(followerUserID, followingUserID uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.FollowStatusAccepted {
		updates["accepted_at"] = time.Now()
	}

	result := r.db.GetConn().Model(&models.UserFollow{}).
		Where("follower_user_id = ? AND following_user_id = ?", followerUserID, followingUserID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("フォロー関係が見つかりません")
	}
	return nil
}

// GetFollowers フォロワー一覧を取得
func (r *userFollowRepository) GetFollowers(userID uuid.UUID) ([]models.UserFollow, error) {
	var follows []models.UserFollow
	err := r.db.GetConn().
		Preload("Follower").
		Where("following_user_id = ? AND status = ?", userID, models.FollowStatusAccepted).
		Order("accepted_at DESC").
		Find(&follows).Error
	return follows, err
}

// GetFollowing フォロー中のユーザー一覧を取得
func (r *userFollowRepository) GetFollowing(userID uuid.UUID) ([]models.UserFollow, error) {
	var follows []models.UserFollow
	err := r.db.GetConn().
		Preload("Following").
		Where("follower_user_id = ? AND status = ?", userID, models.FollowStatusAccepted).
		Order("accepted_at DESC").
		Find(&follows).Error
	return follows, err
}

// GetMutualFriends 相互フォロー（フレンド）一覧を取得
func (r *userFollowRepository) GetMutualFriends(userID uuid.UUID) ([]models.User, error) {
	var friends []models.User

	// サブクエリで相互フォローのユーザーIDを取得
	subQuery := r.db.GetConn().
		Table("user_follows uf1").
		Select("uf1.following_user_id").
		Joins("INNER JOIN user_follows uf2 ON uf1.follower_user_id = uf2.following_user_id AND uf1.following_user_id = uf2.follower_user_id").
		Where("uf1.follower_user_id = ? AND uf1.status = ? AND uf2.status = ?",
			userID, models.FollowStatusAccepted, models.FollowStatusAccepted)

	err := r.db.GetConn().
		Where("id IN (?)", subQuery).
		Find(&friends).Error

	return friends, err
}

// GetFriendCount フレンド数を取得
func (r *userFollowRepository) GetFriendCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.GetConn().
		Table("user_follows uf1").
		Joins("INNER JOIN user_follows uf2 ON uf1.follower_user_id = uf2.following_user_id AND uf1.following_user_id = uf2.follower_user_id").
		Where("uf1.follower_user_id = ? AND uf1.status = ? AND uf2.status = ?",
			userID, models.FollowStatusAccepted, models.FollowStatusAccepted).
		Count(&count).Error

	return count, err
}

// IsMutualFollow 相互フォロー状態かどうかを確認
func (r *userFollowRepository) IsMutualFollow(userID1, userID2 uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetConn().
		Model(&models.UserFollow{}).
		Where("((follower_user_id = ? AND following_user_id = ?) OR (follower_user_id = ? AND following_user_id = ?)) AND status = ?",
			userID1, userID2, userID2, userID1, models.FollowStatusAccepted).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count == 2, nil
}
