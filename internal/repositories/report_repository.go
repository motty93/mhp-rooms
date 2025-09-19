package repositories

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"mhp-rooms/internal/models"
	"gorm.io/gorm"
)

// ReportRepository ユーザー通報関連のリポジトリ
type ReportRepository struct {
	db *gorm.DB
}

// NewReportRepository リポジトリのコンストラクタ
func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Create 通報を作成
func (r *ReportRepository) Create(report *models.UserReport) error {
	return r.db.Create(report).Error
}

// GetByID IDで通報を取得
func (r *ReportRepository) GetByID(id uuid.UUID) (*models.UserReport, error) {
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

// GetByReportedUserID 通報対象ユーザーの通報一覧を取得
func (r *ReportRepository) GetByReportedUserID(userID uuid.UUID, limit int) ([]models.UserReport, error) {
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

// GetByReporterUserID 通報者の通報一覧を取得
func (r *ReportRepository) GetByReporterUserID(userID uuid.UUID, limit int) ([]models.UserReport, error) {
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

// GetPendingReports 未対応の通報一覧を取得
func (r *ReportRepository) GetPendingReports(limit int, offset int) ([]models.UserReport, int64, error) {
	var reports []models.UserReport
	var total int64

	// 合計数を取得
	if err := r.db.Model(&models.UserReport{}).
		Where("status = ?", models.ReportStatusPending).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// データを取得
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

// UpdateStatus 通報のステータスを更新
func (r *ReportRepository) UpdateStatus(id uuid.UUID, status models.ReportStatus, adminNote *string) error {
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

// CheckDuplicateReport 重複通報をチェック（24時間以内の同じ通報者・対象者の組み合わせ）
func (r *ReportRepository) CheckDuplicateReport(reporterID, reportedID uuid.UUID) (bool, error) {
	var count int64
	yesterday := time.Now().Add(-24 * time.Hour)

	err := r.db.Model(&models.UserReport{}).
		Where("reporter_user_id = ? AND reported_user_id = ? AND created_at > ?",
			reporterID, reportedID, yesterday).
		Count(&count).Error

	return count > 0, err
}

// AddAttachment 通報に添付ファイルを追加
func (r *ReportRepository) AddAttachment(attachment *models.ReportAttachment) error {
	return r.db.Create(attachment).Error
}

// GetAttachmentsByReportID 通報IDで添付ファイル一覧を取得
func (r *ReportRepository) GetAttachmentsByReportID(reportID uuid.UUID) ([]models.ReportAttachment, error) {
	var attachments []models.ReportAttachment
	err := r.db.Where("report_id = ?", reportID).Find(&attachments).Error
	return attachments, err
}

// DeleteAttachment 添付ファイルを削除
func (r *ReportRepository) DeleteAttachment(id uuid.UUID) error {
	return r.db.Delete(&models.ReportAttachment{}, "id = ?", id).Error
}

// GetReportStatsByUserID ユーザーの通報統計を取得
func (r *ReportRepository) GetReportStatsByUserID(userID uuid.UUID) (map[string]int64, error) {
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

// SearchReports 通報を検索（管理画面用）
type ReportSearchParams struct {
	Status         *models.ReportStatus
	ReporterID     *uuid.UUID
	ReportedID     *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
	OrderBy        string
	Limit          int
	Offset         int
}

func (r *ReportRepository) SearchReports(params ReportSearchParams) ([]models.UserReport, int64, error) {
	query := r.db.Model(&models.UserReport{})

	// 条件を追加
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

	// 合計数を取得
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ソート順を設定
	if params.OrderBy == "" {
		params.OrderBy = "created_at DESC"
	}

	// データを取得
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

// BatchUpdateStatus 複数の通報のステータスを一括更新
func (r *ReportRepository) BatchUpdateStatus(ids []uuid.UUID, status models.ReportStatus, adminNote *string) error {
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