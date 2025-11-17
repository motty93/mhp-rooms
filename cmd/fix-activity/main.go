package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence"
	"mhp-rooms/internal/models"
)

func main() {
	log.Println("アクティビティデータ修正スクリプトを開始します...")

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

	// room_joinタイプのアクティビティで、descriptionが「ホスト: 」または空のものを取得
	var activities []models.UserActivity
	result := db.GetConn().Where("activity_type = ? AND (description = ? OR description = ? OR description IS NULL)",
		models.ActivityRoomJoin, "ホスト: ", "").Find(&activities)

	if result.Error != nil {
		log.Fatalf("アクティビティの取得に失敗しました: %v", result.Error)
	}

	log.Printf("修正対象のアクティビティ: %d件", len(activities))

	if len(activities) == 0 {
		log.Println("修正対象のアクティビティが見つかりませんでした")
		return
	}

	// 修正処理
	successCount := 0
	failCount := 0

	for _, activity := range activities {
		// metadataからhost_user_idを取得
		if activity.Metadata.Data == nil {
			log.Printf("アクティビティID %s: metadataがnullです", activity.ID)
			failCount++
			continue
		}

		// Metadata.DataをJSON形式にマーシャル後、RoomActivityMetadataにアンマーシャル
		metadataBytes, err := json.Marshal(activity.Metadata.Data)
		if err != nil {
			log.Printf("アクティビティID %s: metadata変換エラー: %v", activity.ID, err)
			failCount++
			continue
		}

		var metadata models.RoomActivityMetadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			log.Printf("アクティビティID %s: metadata解析エラー: %v", activity.ID, err)
			failCount++
			continue
		}

		// host_user_idからUUIDを取得
		hostUserID, err := uuid.Parse(metadata.HostUserID)
		if err != nil {
			log.Printf("アクティビティID %s: host_user_id解析エラー: %v", activity.ID, err)
			failCount++
			continue
		}

		// ホストユーザー情報を取得
		var hostUser models.User
		if err := db.GetConn().Where("id = ?", hostUserID).First(&hostUser).Error; err != nil {
			log.Printf("アクティビティID %s: ホストユーザー取得エラー (ID: %s): %v", activity.ID, hostUserID, err)
			failCount++
			continue
		}

		// DisplayNameが空の場合はUsernameを使用
		hostDisplayName := hostUser.DisplayName
		if hostDisplayName == "" && hostUser.Username != nil {
			hostDisplayName = *hostUser.Username
		}

		if hostDisplayName == "" {
			log.Printf("アクティビティID %s: ホストユーザー名が取得できません (ユーザーID: %s)", activity.ID, hostUserID)
			failCount++
			continue
		}

		// descriptionを更新
		newDescription := fmt.Sprintf("ホスト: %s", hostDisplayName)
		if err := db.GetConn().Model(&activity).Update("description", newDescription).Error; err != nil {
			log.Printf("アクティビティID %s: 更新エラー: %v", activity.ID, err)
			failCount++
			continue
		}

		log.Printf("✓ アクティビティID %s: 「%s」に更新しました", activity.ID, newDescription)
		successCount++
	}

	log.Println("========================================")
	log.Printf("修正完了: 成功 %d件, 失敗 %d件", successCount, failCount)
	log.Println("========================================")
}
