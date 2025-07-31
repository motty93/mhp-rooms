package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"mhp-rooms/internal/database"
	"mhp-rooms/internal/models"
)

func main() {
	// .envファイルの読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .envファイルが見つかりません: %v", err)
	}

	log.Println("パスワード付き部屋のテストデータを作成します...")

	// データベース接続
	log.Println("データベース接続を待機中...")
	db, err := database.ConnectWithRetry()
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}

	// ユーザー取得
	var users []models.User
	if err := db.Limit(3).Find(&users).Error; err != nil {
		log.Fatalf("ユーザーの取得に失敗しました: %v", err)
	}

	// ゲームバージョン取得
	var gameVersions []models.GameVersion
	if err := db.Order("display_order").Find(&gameVersions).Error; err != nil {
		log.Fatalf("ゲームバージョンの取得に失敗しました: %v", err)
	}

	// パスワード付き部屋を作成
	passwordRooms := []struct {
		RoomCode        string
		Name            string
		Description     string
		GameVersionCode string
		Password        string
		TargetMonster   string
		RankRequirement string
		HostIndex       int
	}{
		{
			RoomCode:        "PWD-001",
			Name:            "【鍵】初心者歓迎！",
			Description:     "初心者歓迎のパスワード付き部屋です。パスワード: test123",
			GameVersionCode: "MHP",
			Password:        "test123",
			TargetMonster:   "リオレウス",
			RankRequirement: "制限なし",
			HostIndex:       0,
		},
		{
			RoomCode:        "PWD-002",
			Name:            "【鍵付き】上級者専用",
			Description:     "上級者限定の部屋です。パスワード: advanced456",
			GameVersionCode: "MHP2G",
			Password:        "advanced456",
			TargetMonster:   "ナルガクルガ",
			RankRequirement: "HR7以上",
			HostIndex:       1,
		},
		{
			RoomCode:        "PWD-003",
			Name:            "【鍵】プライベート部屋",
			Description:     "プライベート部屋です。パスワード: private789",
			GameVersionCode: "MHP3",
			Password:        "private789",
			TargetMonster:   "ジンオウガ",
			RankRequirement: "HR5以上",
			HostIndex:       2,
		},
	}

	for _, roomData := range passwordRooms {
		// ゲームバージョンを取得
		var gameVersion models.GameVersion
		for _, gv := range gameVersions {
			if gv.Code == roomData.GameVersionCode {
				gameVersion = gv
				break
			}
		}

		// ルーム作成
		room := models.Room{
			RoomCode:        roomData.RoomCode,
			Name:            roomData.Name,
			Description:     roomData.Description,
			GameVersionID:   gameVersion.ID,
			HostUserID:      users[roomData.HostIndex].ID,
			MaxPlayers:      4,
			CurrentPlayers:  1,
			TargetMonster:   &roomData.TargetMonster,
			RankRequirement: roomData.RankRequirement,
			IsActive:        true,
			IsClosed:        false,
		}

		// パスワードを設定
		if err := room.SetPassword(roomData.Password); err != nil {
			log.Printf("パスワード設定エラー (%s): %v", roomData.RoomCode, err)
			continue
		}

		// ルームを保存
		if err := db.Create(&room).Error; err != nil {
			log.Printf("ルーム作成エラー (%s): %v", roomData.RoomCode, err)
			continue
		}

		log.Printf("✅ パスワード付き部屋を作成しました: %s (パスワード: %s)", roomData.Name, roomData.Password)

		// ルームメンバーを作成
		roomMember := models.RoomMember{
			RoomID:       room.ID,
			UserID:       users[roomData.HostIndex].ID,
			PlayerNumber: 1,
			IsHost:       true,
			Status:       "active",
		}

		if err := db.Create(&roomMember).Error; err != nil {
			log.Printf("ルームメンバー作成エラー (%s): %v", roomData.RoomCode, err)
			continue
		}

		// ルームログを作成
		roomLog := models.RoomLog{
			RoomID: room.ID,
			UserID: &users[roomData.HostIndex].ID,
			Action: "create",
			Details: map[string]interface{}{
				"user_name": users[roomData.HostIndex].DisplayName,
			},
		}

		if err := db.Create(&roomLog).Error; err != nil {
			log.Printf("ルームログ作成エラー (%s): %v", roomData.RoomCode, err)
		}
	}

	log.Println("パスワード付き部屋のテストデータ作成が完了しました")
	os.Exit(0)
}