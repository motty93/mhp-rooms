package repository

import (
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

// platformRepository はプラットフォーム関連の操作を行うリポジトリの実装
type platformRepository struct {
	db *postgres.DB
}

// NewPlatformRepository は新しいPlatformRepositoryインスタンスを作成
func NewPlatformRepository(db *postgres.DB) PlatformRepository {
	return &platformRepository{db: db}
}

// GetActivePlatforms はアクティブなプラットフォーム一覧を取得
func (r *platformRepository) GetActivePlatforms() ([]models.Platform, error) {
	var platforms []models.Platform
	err := r.db.GetConn().
		Order("display_order ASC").
		Find(&platforms).Error

	return platforms, err
}
