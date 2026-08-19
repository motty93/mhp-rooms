package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mhp-rooms/internal/models"
)

// notificationRepository はお知らせ関連の操作を行うリポジトリの実装
type notificationRepository struct {
	db DBInterface
}

// NewNotificationRepository は新しいNotificationRepositoryインスタンスを作成
func NewNotificationRepository(db DBInterface) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create お知らせを作成
func (r *notificationRepository) Create(notification *models.Notification) error {
	if notification == nil {
		return errors.New("お知らせがnilです")
	}
	if notification.UserID == uuid.Nil {
		return errors.New("ユーザーIDが必須です")
	}
	if notification.Type == "" || notification.Title == "" {
		return errors.New("種類とタイトルが必須です")
	}

	return r.db.GetConn().Create(notification).Error
}

// ListByUser ユーザー宛のお知らせを新しい順に取得
func (r *notificationRepository) ListByUser(userID uuid.UUID, limit int) ([]models.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var notifications []models.Notification
	err := r.db.GetConn().
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// CountUnread 未読のお知らせ数を取得
func (r *notificationRepository) CountUnread(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.GetConn().
		Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error

	return count, err
}

// MarkAllRead 未読のお知らせをすべて既読にする
func (r *notificationRepository) MarkAllRead(userID uuid.UUID, readAt time.Time) error {
	return r.db.GetConn().
		Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", readAt).Error
}

// GetState ユーザーの閲覧状態を取得（未作成なら nil）
func (r *notificationRepository) GetState(userID uuid.UUID) (*models.UserNotificationState, error) {
	var state models.UserNotificationState
	err := r.db.GetConn().Where("user_id = ?", userID).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &state, nil
}

// UpsertInfoReadAt 更新情報の既読日時を保存（未作成なら作成）
func (r *notificationRepository) UpsertInfoReadAt(userID uuid.UUID, readAt time.Time) error {
	state := models.UserNotificationState{
		UserID:     userID,
		InfoReadAt: &readAt,
		UpdatedAt:  readAt,
	}

	return r.db.GetConn().
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"info_read_at", "updated_at"}),
		}).
		Create(&state).Error
}
