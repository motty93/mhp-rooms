package main

import (
	"log"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("テストデータの作成を開始します...")

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

	users := []models.User{
		{
			ID:             uuid.New(),
			SupabaseUserID: uuid.New(),
			Email:          "hunter1@example.com",
			Username:       stringPtr("ハンター太郎"),
			DisplayName:    "ハンター太郎",
			AvatarURL:      nil,
			Bio:            stringPtr("MHP2Gメインでプレイしています"),
			IsActive:       true,
			Role:           "user",
		},
		{
			ID:             uuid.New(),
			SupabaseUserID: uuid.New(),
			Email:          "hunter2@example.com",
			Username:       stringPtr("猫好きハンター"),
			DisplayName:    "猫好きハンター",
			AvatarURL:      nil,
			Bio:            stringPtr("初心者歓迎です"),
			IsActive:       true,
			Role:           "user",
		},
		{
			ID:             uuid.New(),
			SupabaseUserID: uuid.New(),
			Email:          "hunter3@example.com",
			Username:       stringPtr("素材コレクター"),
			DisplayName:    "素材コレクター",
			AvatarURL:      nil,
			Bio:            stringPtr("効率重視で狩りをします"),
			IsActive:       true,
			Role:           "user",
		},
	}

	for _, user := range users {
		if err := db.GetConn().Create(&user).Error; err != nil {
			log.Printf("ユーザー作成エラー: %v", err)
		} else {
			log.Printf("ユーザー作成: %s", user.DisplayName)
		}
	}

	var gameVersions []models.GameVersion
	db.GetConn().Find(&gameVersions)

	if len(gameVersions) == 0 {
		log.Fatal("ゲームバージョンが見つかりません。先にマイグレーションを実行してください。")
	}

	rooms := []models.Room{
		{
			ID:              uuid.New(),
			RoomCode:        "ROOM-001",
			Name:            "上位ティガレックス討伐",
			Description:     stringPtr("ティガレックス討伐クエストを一緒にやりませんか？装備自由です"),
			GameVersionID:   gameVersions[2].ID, // MHP2G
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  3,
			TargetMonster:   stringPtr("ティガレックス"),
			RankRequirement: stringPtr("HR6以上"),
			IsActive:        true,
		},
		{
			ID:              uuid.New(),
			RoomCode:        "ROOM-002",
			Name:            "初心者歓迎部屋",
			Description:     stringPtr("下位クエストでゆっくり楽しみましょう！初心者大歓迎です"),
			GameVersionID:   gameVersions[3].ID, // MHP3
			HostUserID:      users[1].ID,
			MaxPlayers:      4,
			CurrentPlayers:  2,
			TargetMonster:   nil,
			RankRequirement: stringPtr("制限なし"),
			IsActive:        true,
		},
		{
			ID:              uuid.New(),
			RoomCode:        "ROOM-003",
			Name:            "レア素材狙い",
			Description:     stringPtr("レア素材狙いで効率よく周回します。経験者推奨"),
			GameVersionID:   gameVersions[1].ID, // MHP2
			HostUserID:      users[2].ID,
			MaxPlayers:      4,
			CurrentPlayers:  4,
			TargetMonster:   stringPtr("リオレウス"),
			RankRequirement: stringPtr("HR4以上"),
			IsActive:        true,
		},
		{
			ID:              uuid.New(),
			RoomCode:        "ROOM-004",
			Name:            "進行中の部屋",
			Description:     stringPtr("現在クエスト中です"),
			GameVersionID:   gameVersions[0].ID, // MHP
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  3,
			TargetMonster:   stringPtr("モノブロス"),
			RankRequirement: stringPtr("HR3以上"),
			IsActive:        true,
		},
	}

	for _, room := range rooms {
		if err := db.GetConn().Create(&room).Error; err != nil {
			log.Printf("ルーム作成エラー: %v", err)
		} else {
			log.Printf("ルーム作成: %s", room.Name)
		}
	}

	log.Println("テストデータの作成が完了しました")
}

func stringPtr(s string) *string {
	return &s
}
