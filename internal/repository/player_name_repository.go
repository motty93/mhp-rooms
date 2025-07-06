package repository

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

type playerNameRepository struct {
	db *postgres.DB
}

func NewPlayerNameRepository(db *postgres.DB) PlayerNameRepository {
	return &playerNameRepository{db: db}
}

// CreatePlayerName 新しいプレイヤー名を作成
func (r *playerNameRepository) CreatePlayerName(playerName *models.PlayerName) error {
	// 同じユーザーとゲームバージョンの組み合わせが既に存在するかチェック
	var existing models.PlayerName
	err := r.db.GetConn().Where("user_id = ? AND game_version_id = ?", playerName.UserID, playerName.GameVersionID).First(&existing).Error
	if err == nil {
		return errors.New("このゲームバージョンのプレイヤー名は既に登録されています")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return r.db.GetConn().Create(playerName).Error
}

// UpdatePlayerName プレイヤー名を更新
func (r *playerNameRepository) UpdatePlayerName(playerName *models.PlayerName) error {
	return r.db.GetConn().Save(playerName).Error
}

// FindPlayerNameByUserAndGame ユーザーIDとゲームバージョンIDでプレイヤー名を取得
func (r *playerNameRepository) FindPlayerNameByUserAndGame(userID, gameVersionID uuid.UUID) (*models.PlayerName, error) {
	var playerName models.PlayerName
	err := r.db.GetConn().
		Preload("User").
		Preload("GameVersion").
		Where("user_id = ? AND game_version_id = ?", userID, gameVersionID).
		First(&playerName).Error
	if err != nil {
		return nil, err
	}
	return &playerName, nil
}

// FindAllPlayerNamesByUser ユーザーIDで全てのプレイヤー名を取得
func (r *playerNameRepository) FindAllPlayerNamesByUser(userID uuid.UUID) ([]models.PlayerName, error) {
	var playerNames []models.PlayerName
	err := r.db.GetConn().
		Preload("GameVersion").
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&playerNames).Error
	return playerNames, err
}

// DeletePlayerName プレイヤー名を削除
func (r *playerNameRepository) DeletePlayerName(id uuid.UUID) error {
	return r.db.GetConn().Delete(&models.PlayerName{}, id).Error
}

// UpsertPlayerName プレイヤー名を作成または更新
func (r *playerNameRepository) UpsertPlayerName(playerName *models.PlayerName) error {
	// 既存のレコードを検索
	var existing models.PlayerName
	err := r.db.GetConn().Where("user_id = ? AND game_version_id = ?", playerName.UserID, playerName.GameVersionID).First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 存在しない場合は新規作成
		return r.db.GetConn().Create(playerName).Error
	}

	if err != nil {
		return err
	}

	// 存在する場合は更新
	existing.Name = playerName.Name
	return r.db.GetConn().Save(&existing).Error
}
