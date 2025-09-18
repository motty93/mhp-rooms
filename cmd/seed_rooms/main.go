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
	log.Println("部屋のテストデータを作成します...")

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

	// ユーザーを取得
	var users []models.User
	db.GetConn().Limit(3).Find(&users)
	if len(users) == 0 {
		log.Fatal("ユーザーが見つかりません。先にユーザーデータを作成してください。")
	}

	// ゲームバージョンを取得
	var gameVersions []models.GameVersion
	db.GetConn().Order("display_order").Find(&gameVersions)
	if len(gameVersions) == 0 {
		log.Fatal("ゲームバージョンが見つかりません。先にマイグレーションを実行してください。")
	}

	// ゲームバージョンをコードでマップ化
	gameVersionMap := make(map[string]models.GameVersion)
	for _, gv := range gameVersions {
		gameVersionMap[gv.Code] = gv
	}

	// パスワード付き部屋を含むテストデータを作成
	rooms := []models.Room{
		// MHP部屋
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP-001",
			Name:            "初心者歓迎！村クエスト",
			Description:     stringPtr("MHP初心者歓迎です！村クエストを一緒に進めましょう"),
			GameVersionID:   gameVersionMap["MHP"].ID,
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   stringPtr("ランポス"),
			RankRequirement: stringPtr("制限なし"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP-002",
			Name:            "リオレウス討伐",
			Description:     stringPtr("リオレウス討伐を手伝ってください"),
			GameVersionID:   gameVersionMap["MHP"].ID,
			HostUserID:      users[1].ID,
			MaxPlayers:      4,
			CurrentPlayers:  2,
			TargetMonster:   stringPtr("リオレウス"),
			RankRequirement: stringPtr("HR3以上"),
			IsActive:        true,
		},
		// MHP2部屋
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP2-001",
			Name:            "素材集め周回",
			Description:     stringPtr("素材集めを効率よく周回します"),
			GameVersionID:   gameVersionMap["MHP2"].ID,
			HostUserID:      users[2].ID,
			MaxPlayers:      4,
			CurrentPlayers:  3,
			TargetMonster:   stringPtr("ティガレックス"),
			RankRequirement: stringPtr("HR4以上"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP2-002",
			Name:            "訓練所でスキル上げ",
			Description:     stringPtr("訓練所でスキルを磨きましょう"),
			GameVersionID:   gameVersionMap["MHP2"].ID,
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   nil,
			RankRequirement: stringPtr("制限なし"),
			IsActive:        true,
		},
		// MHP2G部屋
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP2G-001",
			Name:            "G級ナルガクルガ",
			Description:     stringPtr("G級ナルガクルガ討伐！装備自由"),
			GameVersionID:   gameVersionMap["MHP2G"].ID,
			HostUserID:      users[1].ID,
			MaxPlayers:      4,
			CurrentPlayers:  2,
			TargetMonster:   stringPtr("ナルガクルガ"),
			RankRequirement: stringPtr("HR9以上"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP2G-002",
			Name:            "古龍種連戦",
			Description:     stringPtr("古龍種を順番に討伐していきます"),
			GameVersionID:   gameVersionMap["MHP2G"].ID,
			HostUserID:      users[2].ID,
			MaxPlayers:      4,
			CurrentPlayers:  4,
			TargetMonster:   stringPtr("クシャルダオラ"),
			RankRequirement: stringPtr("HR9以上"),
			IsActive:        true,
		},
		// パスワード付き部屋（MHP2G）
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP2G-003",
			Name:            "【鍵】レア素材狙い専用",
			Description:     stringPtr("効率重視！パスワードはDMで"),
			GameVersionID:   gameVersionMap["MHP2G"].ID,
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   stringPtr("ミラボレアス"),
			RankRequirement: stringPtr("HR9以上"),
			IsActive:        true,
		},
		// MHP3部屋
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP3-001",
			Name:            "温泉チケット集め",
			Description:     stringPtr("温泉チケット集めましょう！"),
			GameVersionID:   gameVersionMap["MHP3"].ID,
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  2,
			TargetMonster:   nil,
			RankRequirement: stringPtr("制限なし"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP3-002",
			Name:            "ジンオウガ狩猟",
			Description:     stringPtr("ジンオウガの素材が欲しいです"),
			GameVersionID:   gameVersionMap["MHP3"].ID,
			HostUserID:      users[1].ID,
			MaxPlayers:      4,
			CurrentPlayers:  3,
			TargetMonster:   stringPtr("ジンオウガ"),
			RankRequirement: stringPtr("HR5以上"),
			IsActive:        true,
		},
		// パスワード付き部屋の例
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHP3-003",
			Name:            "【鍵付き】上級者限定部屋",
			Description:     stringPtr("上級者のみ！パスワード: hunter123"),
			GameVersionID:   gameVersionMap["MHP3"].ID,
			HostUserID:      users[2].ID,
			MaxPlayers:      4,
			CurrentPlayers:  2,
			TargetMonster:   stringPtr("アマツマガツチ"),
			RankRequirement: stringPtr("HR7以上"),
			IsActive:        true,
		},
		// MHXX部屋
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHXX-001",
			Name:            "二つ名持ちモンスター",
			Description:     stringPtr("二つ名持ちモンスターに挑戦！"),
			GameVersionID:   gameVersionMap["MHXX"].ID,
			HostUserID:      users[2].ID,
			MaxPlayers:      4,
			CurrentPlayers:  2,
			TargetMonster:   stringPtr("紅兜アオアシラ"),
			RankRequirement: stringPtr("HR8以上"),
			IsActive:        true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			RoomCode:        "MHXX-002",
			Name:            "G級クエスト進行",
			Description:     stringPtr("G級クエストを順番に進めていきます"),
			GameVersionID:   gameVersionMap["MHXX"].ID,
			HostUserID:      users[0].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   nil,
			RankRequirement: stringPtr("HR8以上"),
			IsActive:        true,
		},
	}

	// パスワードを設定する部屋の情報
	passwordRooms := map[string]string{
		"MHP3-003":  "hunter123",
		"MHP2G-003": "secret456",
	}

	for i, room := range rooms {
		// パスワードがある部屋の場合はパスワードを設定
		if password, exists := passwordRooms[room.RoomCode]; exists {
			if err := room.SetPassword(password); err != nil {
				log.Printf("パスワード設定エラー (%s): %v", room.RoomCode, err)
				continue
			}
			rooms[i] = room // パスワードハッシュを反映
		}

		if err := db.GetConn().Create(&room).Error; err != nil {
			log.Printf("ルーム作成エラー: %v", err)
		} else {
			if room.HasPassword() {
				log.Printf("ルーム作成: %s (%s) - パスワード付き", room.Name, room.RoomCode)
			} else {
				log.Printf("ルーム作成: %s (%s)", room.Name, room.RoomCode)
			}
		}
	}

	log.Println("部屋のテストデータ作成が完了しました")
}

func stringPtr(s string) *string {
	return &s
}
