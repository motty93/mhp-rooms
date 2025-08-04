package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

// userRepository はユーザー関連の操作を行うリポジトリの実装
type userRepository struct {
	db *postgres.DB
}

// NewUserRepository は新しいUserRepositoryインスタンスを作成
func NewUserRepository(db *postgres.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser はユーザーを作成
func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.GetConn().Create(user).Error
}

// FindUserByID はIDでユーザーを検索
func (r *userRepository) FindUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, err
	}
	return &user, nil
}

// FindUsersByIDs は複数のIDでユーザーを一括取得
func (r *userRepository) FindUsersByIDs(ids []uuid.UUID) ([]models.User, error) {
	var users []models.User
	if len(ids) == 0 {
		return users, nil
	}

	err := r.db.GetConn().Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// FindUserBySupabaseUserID はSupabaseユーザーIDでユーザーを検索
func (r *userRepository) FindUserBySupabaseUserID(supabaseUserID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().Where("supabase_user_id = ?", supabaseUserID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail はメールアドレスでユーザーを検索
func (r *userRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.GetConn().Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ユーザーが見つかりません")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser はユーザー情報を更新
func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.GetConn().Save(user).Error
}

// GetActiveUsers はアクティブなユーザー一覧を取得
func (r *userRepository) GetActiveUsers(limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.GetConn().
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}
