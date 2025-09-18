package main

import (
	"log"

	"github.com/joho/godotenv"
	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence"
)

func main() {
	log.Println("マイグレーションコマンドを開始します...")

	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数から設定を読み込みます。")
	}

	config.Init()

	log.Printf("データベース接続を初期化中... (タイプ: %s)", config.AppConfig.Database.Type)
	db, err := persistence.NewDBAdapter(config.AppConfig)
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
