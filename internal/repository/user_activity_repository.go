package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"mhp-rooms/internal/models"
)

// userActivityRepository はユーザーアクティビティ関連の操作を行うリポジトリの実装
type userActivityRepository struct {
	db DBInterface
}

// NewUserActivityRepository は新しいUserActivityRepositoryインスタンスを作成
func NewUserActivityRepository(db DBInterface) UserActivityRepository {
	return &userActivityRepository{db: db}
}

// CreateActivity アクティビティを作成
func (r *userActivityRepository) CreateActivity(activity *models.UserActivity) error {
	if activity == nil {
		return errors.New("アクティビティオブジェクトがnilです")
	}

	// 必須フィールドのバリデーション
	if activity.UserID == uuid.Nil {
		return errors.New("ユーザーIDが必須です")
	}
	if activity.ActivityType == "" {
		return errors.New("アクティビティタイプが必須です")
	}
	if activity.Title == "" {
		return errors.New("タイトルが必須です")
	}

	// 作成日時と更新日時を設定（GORMが自動で設定しない場合のフォールバック）
	now := time.Now()
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = now
	}
	if activity.UpdatedAt.IsZero() {
		activity.UpdatedAt = now
	}

	return r.db.GetConn().Create(activity).Error
}

// GetUserActivities 指定したユーザーのアクティビティを時系列順で取得
func (r *userActivityRepository) GetUserActivities(userID uuid.UUID, limit, offset int) ([]models.UserActivity, error) {
	if userID == uuid.Nil {
		return nil, errors.New("ユーザーIDが必須です")
	}

	var activities []models.UserActivity

	// limitの上限を設定（パフォーマンス保護）
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	err := r.db.GetConn().
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	if err != nil {
		return nil, err
	}

	return activities, nil
}

// GetUserActivitiesByType 指定したユーザーの特定タイプのアクティビティを取得
func (r *userActivityRepository) GetUserActivitiesByType(userID uuid.UUID, activityType string, limit, offset int) ([]models.UserActivity, error) {
	if userID == uuid.Nil {
		return nil, errors.New("ユーザーIDが必須です")
	}
	if activityType == "" {
		return nil, errors.New("アクティビティタイプが必須です")
	}

	var activities []models.UserActivity

	// limitの上限を設定
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	err := r.db.GetConn().
		Where("user_id = ? AND activity_type = ?", userID, activityType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	if err != nil {
		return nil, err
	}

	return activities, nil
}

// CountUserActivities 指定したユーザーのアクティビティ総数を取得
func (r *userActivityRepository) CountUserActivities(userID uuid.UUID) (int64, error) {
	if userID == uuid.Nil {
		return 0, errors.New("ユーザーIDが必須です")
	}

	var count int64
	err := r.db.GetConn().
		Model(&models.UserActivity{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

// DeleteActivity 指定したアクティビティを削除
func (r *userActivityRepository) DeleteActivity(id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("アクティビティIDが必須です")
	}

	result := r.db.GetConn().
		Where("id = ?", id).
		Delete(&models.UserActivity{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("削除対象のアクティビティが見つかりません")
	}

	return nil
}

// DeleteOldActivities 指定した日時より古いアクティビティを削除（データベースメンテナンス用）
func (r *userActivityRepository) DeleteOldActivities(olderThan time.Time) error {
	if olderThan.IsZero() {
		return errors.New("削除基準日時が必須です")
	}

	result := r.db.GetConn().
		Where("created_at < ?", olderThan).
		Delete(&models.UserActivity{})

	return result.Error
}
