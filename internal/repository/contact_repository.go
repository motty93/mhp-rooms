package repository

import (
	"mhp-rooms/internal/models"

	"github.com/google/uuid"
)

// ContactRepository お問合せリポジトリのインターフェース
type ContactRepository interface {
	CreateContact(contact *models.Contact) error
	FindContactByID(id uuid.UUID) (*models.Contact, error)
}

type contactRepository struct {
	db DBInterface
}

// NewContactRepository お問合せリポジトリのコンストラクタ
func NewContactRepository(db DBInterface) ContactRepository {
	return &contactRepository{db: db}
}

// CreateContact お問合せを作成
func (r *contactRepository) CreateContact(contact *models.Contact) error {
	return r.db.GetConn().Create(contact).Error
}

// FindContactByID IDでお問合せを検索
func (r *contactRepository) FindContactByID(id uuid.UUID) (*models.Contact, error) {
	var contact models.Contact
	if err := r.db.GetConn().First(&contact, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &contact, nil
}
