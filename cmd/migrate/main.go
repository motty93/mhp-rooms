package main

import (
	"log"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/database"
)

func main() {
	log.Println("マイグレーションコマンドを開始します...")

	// 設定を初期化
	config.Init()

	log.Println("データベース接続を待機中...")
	if err := database.WaitForDB(config.AppConfig, 30, 2*time.Second); err != nil {
		log.Fatalf("データベース接続待機に失敗しました: %v", err)
	}

	// データベース接続を作成
	log.Println("データベース接続を初期化中...")
	db, err := database.NewDB(config.AppConfig)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	// マイグレーション実行
	log.Println("データベースマイグレーションを実行中...")
	if err := db.Migrate(); err != nil {
		log.Fatalf("マイグレーションに失敗しました: %v", err)
	}

	log.Println("マイグレーションが正常に完了しました")
}