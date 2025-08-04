package main

import (
	"fmt"
	"log"
	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数から設定を読み込みます。")
	}

	config.Init()

	if err := postgres.WaitForDB(config.AppConfig, 30, 2*time.Second); err != nil {
		log.Fatalf("データベース接続待機に失敗しました: %v", err)
	}

	db, err := postgres.NewDB(config.AppConfig)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	// プラットフォーム確認
	var platforms []models.Platform
	db.GetConn().Order("display_order").Find(&platforms)

	fmt.Println("=== プラットフォーム ===")
	for _, p := range platforms {
		fmt.Printf("ID: %s, Name: %s, DisplayOrder: %d\n", p.ID, p.Name, p.DisplayOrder)
	}

	// ゲームバージョン確認
	var gameVersions []models.GameVersion
	db.GetConn().Order("display_order").Find(&gameVersions)

	fmt.Println("\n=== ゲームバージョン ===")
	for _, gv := range gameVersions {
		fmt.Printf("ID: %s, Code: %s, Name: %s, DisplayOrder: %d, PlatformID: %s\n",
			gv.ID, gv.Code, gv.Name, gv.DisplayOrder, gv.PlatformID)
	}
}
