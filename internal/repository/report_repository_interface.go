package repository

import (
	"github.com/google/uuid"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repositories"
	"gorm.io/gorm"
)

// ReportRepositoryInterface 通報リポジトリのインターフェース
type ReportRepositoryInterface interface {
	Create(report *models.UserReport) error
	GetByID(id uuid.UUID) (*models.UserReport, error)
	GetByReportedUserID(userID uuid.UUID, limit int) ([]models.UserReport, error)
	GetByReporterUserID(userID uuid.UUID, limit int) ([]models.UserReport, error)
	GetPendingReports(limit int, offset int) ([]models.UserReport, int64, error)
	UpdateStatus(id uuid.UUID, status models.ReportStatus, adminNote *string) error
	CheckDuplicateReport(reporterID, reportedID uuid.UUID) (bool, error)
	AddAttachment(attachment *models.ReportAttachment) error
	GetAttachmentsByReportID(reportID uuid.UUID) ([]models.ReportAttachment, error)
	DeleteAttachment(id uuid.UUID) error
	GetReportStatsByUserID(userID uuid.UUID) (map[string]int64, error)
	SearchReports(params repositories.ReportSearchParams) ([]models.UserReport, int64, error)
	BatchUpdateStatus(ids []uuid.UUID, status models.ReportStatus, adminNote *string) error
}

// NewReportRepository リポジトリのコンストラクタ
func NewReportRepository(db interface{}) ReportRepositoryInterface {
	return repositories.NewReportRepository(db.(*gorm.DB))
}