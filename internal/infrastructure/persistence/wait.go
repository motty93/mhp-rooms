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

	// Tursoの場合は待機不要（サーバーレスで常時起動）
	// 接続失敗は設定ミスかトークン期限切れなのでリトライしても解決しない
	if cfg.Database.Type == "turso" {
		log.Println("Tursoデータベースに接続中...")
		adapter, err := NewDBAdapter(cfg)
		if err != nil {
			return fmt.Errorf("Tursoデータベース接続エラー: %w", err)
		}
		adapter.Close()
		log.Println("Tursoデータベース接続に成功しました")
		return nil
	}

	// PostgreSQLの場合は接続確認（Docker起動待ち）
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
