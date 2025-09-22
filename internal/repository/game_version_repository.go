package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
)

type gameVersionRepository struct {
	db DBInterface
}

func NewGameVersionRepository(db DBInterface) GameVersionRepository {
	return &gameVersionRepository{db: db}
}

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

func (r *gameVersionRepository) GetActiveGameVersions() ([]models.GameVersion, error) {
	var versions []models.GameVersion
	err := r.db.GetConn().
		Preload("Platform").
		Where("is_active = ?", true).
		Order("display_order ASC").
		Find(&versions).Error

	return versions, err
}
