package persistence

import (
	"fmt"
	"log"
	"time"

	"mhp-rooms/internal/config"
)

// WaitForDB はデータベースが利用可能になるまで待機
func WaitForDB(cfg *config.Config, maxAttempts int, delay time.Duration) error {
	// SQLiteの場合は待機不要
	if cfg.Database.Type == "sqlite" {
		log.Println("SQLiteデータベースを使用します（待機不要）")
		return nil
	}

	// PostgreSQLの場合は接続確認
	for i := 0; i < maxAttempts; i++ {
		adapter, err := NewDBAdapter(cfg)
		if err == nil {
			adapter.Close()
			log.Println("データベース接続に成功しました")
			return nil
		}

		if i < maxAttempts-1 {
			log.Printf("データベース接続待機中... (%d/%d)\n", i+1, maxAttempts)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("データベース接続がタイムアウトしました")
}
