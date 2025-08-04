package main

import (
	"fmt"
	"log"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("パスワード付き部屋のテストを開始します...")

	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数から設定を読み込みます。")
	}

	config.Init()

	log.Println("データベース接続を待機中...")
	if err := postgres.WaitForDB(config.AppConfig, 30, 2*time.Second); err != nil {
		log.Fatalf("データベース接続待機に失敗しました: %v", err)
	}

	db, err := postgres.NewDB(config.AppConfig)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	// パスワード付き部屋を取得
	var rooms []models.Room
	db.GetConn().Where("password_hash IS NOT NULL").Find(&rooms)

	testPasswords := map[string]string{
		"MHP3-003":  "hunter123",
		"MHP2G-003": "secret456",
		"ROOM-005":  "test1234",
	}

	for _, room := range rooms {
		fmt.Printf("\n部屋: %s (%s)\n", room.Name, room.RoomCode)
		fmt.Printf("パスワード保護: %v\n", room.HasPassword())

		if correctPassword, exists := testPasswords[room.RoomCode]; exists {
			// 正しいパスワードでテスト
			if room.CheckPassword(correctPassword) {
				fmt.Printf("✅ 正しいパスワード '%s' で認証成功\n", correctPassword)
			} else {
				fmt.Printf("❌ 正しいパスワード '%s' で認証失敗\n", correctPassword)
			}

			// 間違ったパスワードでテスト
			wrongPassword := "wrong123"
			if room.CheckPassword(wrongPassword) {
				fmt.Printf("❌ 間違ったパスワード '%s' で認証成功（これは問題です）\n", wrongPassword)
			} else {
				fmt.Printf("✅ 間違ったパスワード '%s' で認証失敗（正常）\n", wrongPassword)
			}
		}
	}

	log.Println("\nパスワード付き部屋のテストが完了しました")
}
