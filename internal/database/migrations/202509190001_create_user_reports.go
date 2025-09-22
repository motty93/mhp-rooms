package migrations

import (
	"github.com/motty93/mhp-rooms/internal/models"
	"gorm.io/gorm"
)

// CreateUserReportsTables ユーザー通報関連のテーブルを作成
func CreateUserReportsTables(db *gorm.DB) error {
	// user_reportsテーブル作成
	if err := db.AutoMigrate(&models.UserReport{}); err != nil {
		return err
	}

	// report_attachmentsテーブル作成
	if err := db.AutoMigrate(&models.ReportAttachment{}); err != nil {
		return err
	}

	// インデックス作成
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_user_reports_status ON user_reports(status);
		CREATE INDEX IF NOT EXISTS idx_user_reports_created_at ON user_reports(created_at);
		CREATE INDEX IF NOT EXISTS idx_user_reports_reporter_reported ON user_reports(reporter_user_id, reported_user_id);
	`).Error; err != nil {
		return err
	}

	return nil
}
