package main

import (
	"flag"
	"log"
	"os"

	"mhp-rooms/internal/database"
)

func main() {
	var migrate = flag.Bool("migrate", false, "マイグレーションを実行")
	var rollback = flag.Bool("rollback", false, "ロールバックを実行")
	flag.Parse()

	// データベース接続を初期化
	if err := database.InitDB(); err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer database.CloseDB()

	if *migrate {
		log.Println("マイグレーションを開始します...")
		if err := database.Migrate(); err != nil {
			log.Fatalf("マイグレーションに失敗しました: %v", err)
		}
		log.Println("マイグレーションが完了しました")
		return
	}

	if *rollback {
		log.Println("ロールバック機能は現在実装されていません")
		os.Exit(1)
	}

	// フラグが指定されていない場合はヘルプを表示
	flag.Usage()
}