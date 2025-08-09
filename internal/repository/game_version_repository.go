package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

// gameVersionRepository はゲームバージョン関連の操作を行うリポジトリの実装
type gameVersionRepository struct {
	db *postgres.DB
}

// NewGameVersionRepository は新しいGameVersionRepositoryインスタンスを作成
func NewGameVersionRepository(db *postgres.DB) GameVersionRepository {
	return &gameVersionRepository{db: db}
}

// FindGameVersionByID はIDでゲームバージョンを検索
func (r *gameVersionRepository) FindGameVersionByID(id uuid.UUID) (*models.GameVersion, error) {
	var gameVersion models.GameVersion
	err := r.db.GetConn().Where("id = ?", id).First(&gameVersion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ゲームバージョンが見つかりません")
		}
		return nil, err
	}
	return &gameVersion, nil
}

// FindGameVersionByCode はコードでゲームバージョンを検索
func (r *gameVersionRepository) FindGameVersionByCode(code string) (*models.GameVersion, error) {
	var gameVersion models.GameVersion
	err := r.db.GetConn().Where("code = ?", code).First(&gameVersion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ゲームバージョンが見つかりません")
		}
		return nil, err
	}
	return &gameVersion, nil
}

// GetActiveGameVersions はアクティブなゲームバージョン一覧を取得
func (r *gameVersionRepository) GetActiveGameVersions() ([]models.GameVersion, error) {
	var versions []models.GameVersion
	err := r.db.GetConn().
		Preload("Platform").
		Where("is_active = ?", true).
		Order("display_order ASC").
		Find(&versions).Error

	return versions, err
}
