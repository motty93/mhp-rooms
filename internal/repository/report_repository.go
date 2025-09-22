package repository

import (
	"fmt"
	"time"

	"mhp-rooms/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportSearchParams struct {
	Status     *models.ReportStatus
	ReporterID *uuid.UUID
	ReportedID *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	OrderBy    string
	Limit      int
	Offset     int
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db DBInterface) ReportRepository {
	return &reportRepository{db: db.GetConn()}
}

func (r *reportRepository) Create(report *models.UserReport) error {
	return r.db.Create(report).Error
}

func (r *reportRepository) GetByID(id uuid.UUID) (*models.UserReport, error) {
	var report models.UserReport
	err := r.db.Preload("Reporter").
		Preload("Reported").
		Preload("Attachments").
		First(&report, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) GetByReportedUserID(userID uuid.UUID, limit int) ([]models.UserReport, error) {
	var reports []models.UserReport
	query := r.db.Preload("Reporter").
		Where("reported_user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&reports).Error
	return reports, err
}

func (r *reportRepository) GetByReporterUserID(userID uuid.UUID, limit int) ([]models.UserReport, error) {
	var reports []models.UserReport
	query := r.db.Preload("Reported").
		Where("reporter_user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&reports).Error
	return reports, err
}

func (r *reportRepository) GetPendingReports(limit int, offset int) ([]models.UserReport, int64, error) {
	var reports []models.UserReport
	var total int64

	if err := r.db.Model(&models.UserReport{}).
		Where("status = ?", models.ReportStatusPending).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Reporter").
		Preload("Reported").
		Preload("Attachments").
		Where("status = ?", models.ReportStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	return reports, total, err
}

func (r *reportRepository) UpdateStatus(id uuid.UUID, status models.ReportStatus, adminNote *string) error {
	updates := map[string]interface{}{
		"status":     status,
		"admin_note": adminNote,
	}

	// 解決済みまたは却下の場合は解決日時を記録
	if status == models.ReportStatusResolved || status == models.ReportStatusRejected {
		now := time.Now()
		updates["resolved_at"] = &now
	}

	return r.db.Model(&models.UserReport{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// 重複通報をチェック（24時間以内の同じ通報者・対象者の組み合わせ）
func (r *reportRepository) CheckDuplicateReport(reporterID, reportedID uuid.UUID) (bool, error) {
	var count int64
	yesterday := time.Now().Add(-24 * time.Hour)

	err := r.db.Model(&models.UserReport{}).
		Where("reporter_user_id = ? AND reported_user_id = ? AND created_at > ?",
			reporterID, reportedID, yesterday).
		Count(&count).Error

	return count > 0, err
}

func (r *reportRepository) AddAttachment(attachment *models.ReportAttachment) error {
	return r.db.Create(attachment).Error
}

func (r *reportRepository) GetAttachmentsByReportID(reportID uuid.UUID) ([]models.ReportAttachment, error) {
	var attachments []models.ReportAttachment
	err := r.db.Where("report_id = ?", reportID).Find(&attachments).Error
	return attachments, err
}

func (r *reportRepository) DeleteAttachment(id uuid.UUID) error {
	return r.db.Delete(&models.ReportAttachment{}, "id = ?", id).Error
}

func (r *reportRepository) GetReportStatsByUserID(userID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 通報された回数
	var reportedCount int64
	if err := r.db.Model(&models.UserReport{}).
		Where("reported_user_id = ?", userID).
		Count(&reportedCount).Error; err != nil {
		return nil, err
	}
	stats["reported_count"] = reportedCount

	// 解決済みの通報数
	var resolvedCount int64
	if err := r.db.Model(&models.UserReport{}).
		Where("reported_user_id = ? AND status = ?", userID, models.ReportStatusResolved).
		Count(&resolvedCount).Error; err != nil {
		return nil, err
	}
	stats["resolved_count"] = resolvedCount

	// 却下された通報数
	var rejectedCount int64
	if err := r.db.Model(&models.UserReport{}).
		Where("reported_user_id = ? AND status = ?", userID, models.ReportStatusRejected).
		Count(&rejectedCount).Error; err != nil {
		return nil, err
	}
	stats["rejected_count"] = rejectedCount

	return stats, nil
}

func (r *reportRepository) SearchReports(params ReportSearchParams) ([]models.UserReport, int64, error) {
	query := r.db.Model(&models.UserReport{})

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.ReporterID != nil {
		query = query.Where("reporter_user_id = ?", *params.ReporterID)
	}
	if params.ReportedID != nil {
		query = query.Where("reported_user_id = ?", *params.ReportedID)
	}
	if params.StartDate != nil {
		query = query.Where("created_at >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("created_at <= ?", *params.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.OrderBy == "" {
		params.OrderBy = "created_at DESC"
	}

	var reports []models.UserReport
	err := query.
		Preload("Reporter").
		Preload("Reported").
		Preload("Attachments").
		Order(params.OrderBy).
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&reports).Error

	return reports, total, err
}

// 複数の通報のステータスを一括更新
func (r *reportRepository) BatchUpdateStatus(ids []uuid.UUID, status models.ReportStatus, adminNote *string) error {
	if len(ids) == 0 {
		return fmt.Errorf("更新対象のIDが指定されていません")
	}

	updates := map[string]interface{}{
		"status":     status,
		"admin_note": adminNote,
	}

	// 解決済みまたは却下の場合は解決日時を記録
	if status == models.ReportStatusResolved || status == models.ReportStatusRejected {
		now := time.Now()
		updates["resolved_at"] = &now
	}

	return r.db.Model(&models.UserReport{}).
		Where("id IN ?", ids).
		Updates(updates).Error
}
