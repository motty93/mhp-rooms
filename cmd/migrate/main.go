package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
)

func main() {
	log.Println("マイグレーションコマンドを開始します...")

	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数から設定を読み込みます。")
	}

	config.Init()

	log.Println("データベース接続を待機中...")
	if err := postgres.WaitForDB(config.AppConfig, 30, 2*time.Second); err != nil {
		log.Fatalf("データベース接続待機に失敗しました: %v", err)
	}

	log.Println("データベース接続を初期化中...")
	db, err := postgres.NewDB(config.AppConfig)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	log.Println("データベースマイグレーションを実行中...")
	if err := db.Migrate(); err != nil {
		log.Fatalf("マイグレーションに失敗しました: %v", err)
	}

	log.Println("マイグレーションが正常に完了しました")
}
