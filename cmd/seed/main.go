package main

import (
	"log"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence"
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
	if err := persistence.WaitForDB(config.AppConfig, 30, 2*time.Second); err != nil {
		log.Fatalf("データベース接続待機に失敗しました: %v", err)
	}

	db, err := persistence.NewDBAdapter(config.AppConfig)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	users := []models.User{
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter1@example.com",
			Username:       stringPtr("hunter_taro"),
			DisplayName:    "ハンター太郎",
			AvatarURL:      nil,
			Bio:            stringPtr("MHP2Gメインでプレイしています"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter2@example.com",
			Username:       stringPtr("neko_hunter"),
			DisplayName:    "猫好きハンター🐱",
			AvatarURL:      nil,
			Bio:            stringPtr("初心者歓迎です"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter3@example.com",
			Username:       stringPtr("material_collector"),
			DisplayName:    "素材コレクター",
			AvatarURL:      nil,
			Bio:            stringPtr("効率重視で狩りをします"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter4@example.com",
			Username:       stringPtr("speed_runner"),
			DisplayName:    "スピードランナー",
			AvatarURL:      nil,
			Bio:            stringPtr("タイムアタック専門です"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter5@example.com",
			Username:       stringPtr("casual_gamer"),
			DisplayName:    "まったりハンター",
			AvatarURL:      nil,
			Bio:            stringPtr("のんびりやってます"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter6@example.com",
			Username:       stringPtr("pro_hunter"),
			DisplayName:    "プロハンター",
			AvatarURL:      nil,
			Bio:            stringPtr("攻略情報を共有します"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter7@example.com",
			Username:       stringPtr("weapon_master"),
			DisplayName:    "武器マスター",
			AvatarURL:      nil,
			Bio:            stringPtr("全武器使えます"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter8@example.com",
			Username:       stringPtr("monster_scholar"),
			DisplayName:    "モンスター博士",
			AvatarURL:      nil,
			Bio:            stringPtr("モンスターの生態研究中"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter9@example.com",
			Username:       stringPtr("item_vendor"),
			DisplayName:    "アイテム商人",
			AvatarURL:      nil,
			Bio:            stringPtr("レアアイテム持ってます"),
			IsActive:       true,
			Role:           "user",
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			SupabaseUserID: uuid.New(),
			Email:          "hunter10@example.com",
			Username:       stringPtr("team_leader"),
			DisplayName:    "チームリーダー",
			AvatarURL:      nil,
			Bio:            stringPtr("チーム戦略が得意です"),
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
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "ROOM-001",
			Name:            "上位ティガレックス討伐",
			Description:     stringPtr("ティガレックス討伐クエストを一緒にやりませんか？装備自由です"),
			GameVersionID:   gameVersions[2].ID, // MHP2G
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   stringPtr("ティガレックス"),
			RankRequirement: stringPtr("HR6以上"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "ROOM-002",
			Name:            "初心者歓迎部屋",
			Description:     stringPtr("下位クエストでゆっくり楽しみましょう！初心者大歓迎です"),
			GameVersionID:   gameVersions[3].ID, // MHP3
			HostUserID:      users[1].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   nil,
			RankRequirement: stringPtr("制限なし"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "ROOM-003",
			Name:            "レア素材狙い",
			Description:     stringPtr("レア素材狙いで効率よく周回します。経験者推奨"),
			GameVersionID:   gameVersions[1].ID, // MHP2
			HostUserID:      users[2].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   stringPtr("リオレウス"),
			RankRequirement: stringPtr("HR4以上"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "ROOM-004",
			Name:            "進行中の部屋",
			Description:     stringPtr("現在クエスト中です"),
			GameVersionID:   gameVersions[0].ID, // MHP
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   stringPtr("モノブロス"),
			RankRequirement: stringPtr("HR3以上"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "ROOM-005",
			Name:            "【鍵付き】プライベート部屋",
			Description:     stringPtr("友達限定です。パスワード: test1234"),
			GameVersionID:   gameVersions[2].ID, // MHP2G
			HostUserID:      users[1].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   stringPtr("ラージャン"),
			RankRequirement: stringPtr("HR7以上"),
			IsActive:        true,
		},
	}

	// パスワード付き部屋の設定
	passwordRoom := "ROOM-005"
	password := "test1234"

	for i, room := range rooms {
		// パスワード付き部屋の処理
		if room.RoomCode == passwordRoom {
			if err := room.SetPassword(password); err != nil {
				log.Printf("パスワード設定エラー (%s): %v", room.RoomCode, err)
				continue
			}
			rooms[i] = room // パスワードハッシュを反映
		}

		if err := db.GetConn().Create(&room).Error; err != nil {
			log.Printf("ルーム作成エラー: %v", err)
		} else {
			// ホストをroom_membersテーブルに追加
			member := models.RoomMember{
				ID:           uuid.New(),
				RoomID:       room.ID,
				UserID:       room.HostUserID,
				PlayerNumber: 1,
				IsHost:       true,
				Status:       "active",
				JoinedAt:     time.Now(),
			}
			if err := db.GetConn().Create(&member).Error; err != nil {
				log.Printf("ホストメンバー追加エラー (%s): %v", room.Name, err)
			}

			if room.HasPassword() {
				log.Printf("ルーム作成: %s - パスワード付き", room.Name)
			} else {
				log.Printf("ルーム作成: %s", room.Name)
			}
		}
	}

	log.Println("テストデータの作成が完了しました")
}

func stringPtr(s string) *string {
	return &s
}
