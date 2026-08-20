package repository

import (
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
)

type roomLogRepository struct {
	db DBInterface
}

func NewRoomLogRepository(db DBInterface) RoomLogRepository {
	return &roomLogRepository{db: db}
}

func (r *roomLogRepository) CreateLog(log *models.RoomLog) error {
	return r.db.GetConn().Create(log).Error
}

// ListRecentLogs 全部屋の操作ログを新しい順で取得（管理画面のタイムライン用）
func (r *roomLogRepository) ListRecentLogs(limit, offset int) ([]models.RoomLog, error) {
	var logs []models.RoomLog
	err := r.db.GetConn().
		Preload("Room", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "display_name")
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

func (r *roomLogRepository) CountLogs() (int64, error) {
	var count int64
	err := r.db.GetConn().Model(&models.RoomLog{}).Count(&count).Error
	return count, err
}
