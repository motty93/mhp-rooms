// room-cleanup は一定期間活動がない部屋を自動解散する Cloud Run Job 用コマンド。
// Cloud Scheduler から定期的に実行される想定（詳細は docs/deploy.md を参照）。
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/services"
)

const defaultInactiveHours = 48

func main() {
	startTime := time.Now()

	// .envファイルのロード（ローカル実行用。Cloud Run では環境変数を使用）
	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数を使用します。")
	}

	idleDuration, err := parseInactiveHours(os.Getenv("ROOM_INACTIVE_HOURS"))
	if err != nil {
		log.Fatalf("環境変数 ROOM_INACTIVE_HOURS が不正です: %v", err)
	}
	dryRun := parseBool(os.Getenv("DRY_RUN"))

	log.Printf("部屋の自動削除を開始: idle=%s, dry_run=%t", idleDuration, dryRun)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:           config.GetEnv("DB_TYPE", "turso"),
			TursoURL:       os.Getenv("TURSO_DATABASE_URL"),
			TursoAuthToken: os.Getenv("TURSO_AUTH_TOKEN"),
		},
	}

	dbAdapter, err := persistence.NewDBAdapter(cfg)
	if err != nil {
		log.Fatalf("データベース接続失敗: %v", err)
	}
	defer dbAdapter.Close()

	cleanup := services.NewRoomCleanupService(repository.NewRepository(dbAdapter))

	if dryRun {
		rooms, err := cleanup.FindInactiveRooms(idleDuration)
		if err != nil {
			log.Fatalf("非アクティブな部屋の取得失敗: %v", err)
		}
		for _, room := range rooms {
			log.Printf("[dry-run] 削除対象: room_id=%s name=%q created_at=%s updated_at=%s", room.ID, room.Name, room.CreatedAt.Format(time.RFC3339), room.UpdatedAt.Format(time.RFC3339))
		}
		log.Printf("[dry-run] 削除対象 %d 件（実際の削除は行っていません） duration_ms=%d", len(rooms), time.Since(startTime).Milliseconds())
		return
	}

	dismissed, err := cleanup.DismissInactiveRooms(idleDuration)
	for _, room := range dismissed {
		log.Printf("自動削除: room_id=%s name=%q host_user_id=%s", room.ID, room.Name, room.HostUserID)
	}
	log.Printf("部屋の自動削除完了: dismissed=%d duration_ms=%d", len(dismissed), time.Since(startTime).Milliseconds())

	if err != nil {
		log.Fatalf("一部の部屋の削除に失敗しました: %v", err)
	}
}

// parseInactiveHours ROOM_INACTIVE_HOURS（時間）を Duration に変換する。未指定は defaultInactiveHours
func parseInactiveHours(value string) (time.Duration, error) {
	if value == "" {
		return defaultInactiveHours * time.Hour, nil
	}

	hours, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %q as integer: %w", value, err)
	}
	if hours <= 0 {
		return 0, errors.New("must be a positive integer")
	}

	return time.Duration(hours) * time.Hour, nil
}

// parseBool "true" / "1" を真として扱う
func parseBool(value string) bool {
	return value == "true" || value == "1"
}
