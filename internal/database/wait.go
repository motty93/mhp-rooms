package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mhp-rooms/internal/config"
)

// WaitForDB データベースが利用可能になるまで待機
func WaitForDB(cfg *config.Config, maxRetries int, retryInterval time.Duration) error {
	dsn := cfg.GetDSN()

	for i := 0; i < maxRetries; i++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			// 接続テスト
			sqlDB, err := db.DB()
			if err == nil {
				if err := sqlDB.Ping(); err == nil {
					sqlDB.Close()
					log.Printf("データベース接続が確認できました (試行 %d/%d)", i+1, maxRetries)
					return nil
				}
				sqlDB.Close()
			}
		}

		if i < maxRetries-1 {
			log.Printf("データベース接続待機中... (試行 %d/%d) - %v", i+1, maxRetries, err)
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("データベース接続のタイムアウト: %d回試行後も接続できませんでした", maxRetries)
}
