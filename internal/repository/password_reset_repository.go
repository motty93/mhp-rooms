package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

// PasswordResetRepository はパスワードリセット関連の操作を行うリポジトリのインターface
type PasswordResetRepository interface {
	CreatePasswordReset(reset *models.PasswordReset) error
	FindPasswordResetByToken(token string) (*models.PasswordReset, error)
	MarkPasswordResetAsUsed(id uuid.UUID) error
	DeleteExpiredPasswordResets() error
}

// passwordResetRepository はPasswordResetRepositoryの実装
type passwordResetRepository struct {
	db *postgres.DB
}

// NewPasswordResetRepository は新しいPasswordResetRepositoryインスタンスを作成
func NewPasswordResetRepository(db *postgres.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

// CreatePasswordReset はパスワードリセットレコードを作成
func (r *passwordResetRepository) CreatePasswordReset(reset *models.PasswordReset) error {
	return r.db.GetConn().Create(reset).Error
}

// FindPasswordResetByToken はトークンでパスワードリセットレコードを検索
func (r *passwordResetRepository) FindPasswordResetByToken(token string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	err := r.db.GetConn().Where("token = ? AND used = false AND expires_at > ?", token, time.Now()).
		First(&reset).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("有効なリセットトークンが見つかりません")
		}
		return nil, err
	}
	return &reset, nil
}

// MarkPasswordResetAsUsed はパスワードリセットを使用済みとしてマーク
func (r *passwordResetRepository) MarkPasswordResetAsUsed(id uuid.UUID) error {
	return r.db.GetConn().Model(&models.PasswordReset{}).
		Where("id = ?", id).
		Update("used", true).Error
}

// DeleteExpiredPasswordResets は期限切れのパスワードリセットレコードを削除
func (r *passwordResetRepository) DeleteExpiredPasswordResets() error {
	return r.db.GetConn().Where("expires_at < ?", time.Now()).
		Delete(&models.PasswordReset{}).Error
}
