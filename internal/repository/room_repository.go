package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/models"
)

type roomRepository struct {
	db *postgres.DB
}

func NewRoomRepository(db *postgres.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) CreateRoom(room *models.Room) error {
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		// 部屋を作成
		if err := tx.Create(room).Error; err != nil {
			return err
		}

		// ホストユーザーをメンバーとして追加
		member := models.RoomMember{
			RoomID:       room.ID,
			UserID:       room.HostUserID,
			PlayerNumber: 1,
			IsHost:       true,
			Status:       "active",
			JoinedAt:     time.Now(),
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		// 部屋作成ログを記録
		log := models.RoomLog{
			RoomID: room.ID,
			UserID: &room.HostUserID,
			Action: "create",
			Details: models.JSONB{
				"room_name": room.Name,
			},
		}
		return tx.Create(&log).Error
	})
}

func (r *roomRepository) FindRoomByID(id uuid.UUID) (*models.Room, error) {
	var room models.Room
	err := r.db.GetConn().
		Preload("GameVersion").
		Preload("Host").
		Where("id = ?", id).
		First(&room).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ルームが見つかりません")
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) FindRoomByRoomCode(roomCode string) (*models.Room, error) {
	var room models.Room
	err := r.db.GetConn().
		Preload("GameVersion").
		Preload("Host").
		Where("room_code = ?", roomCode).
		First(&room).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ルームが見つかりません")
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error) {
	var rooms []models.Room
	query := r.db.GetConn().
		Select("rooms.*, COUNT(room_members.id) as current_players").
		Joins("LEFT JOIN room_members ON rooms.id = room_members.room_id AND room_members.status = 'active'").
		Preload("GameVersion").
		Preload("Host").
		Where("rooms.is_active = ?", true).
		Group("rooms.id")

	if gameVersionID != nil {
		query = query.Where("rooms.game_version_id = ?", *gameVersionID)
	}

	err := query.
		Order("rooms.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rooms).Error
	return rooms, err
}

// GetActiveRoomsWithJoinStatus ユーザーの参加状態を含めて部屋一覧を取得（パフォーマンス最適化版）
func (r *roomRepository) GetActiveRoomsWithJoinStatus(userID *uuid.UUID, gameVersionID *uuid.UUID, limit, offset int) ([]models.RoomWithJoinStatus, error) {
	if userID == nil {
		// ユーザーIDがnilの場合は、通常の部屋一覧を取得してisJoinedをfalseに設定
		normalRooms, err := r.GetActiveRooms(gameVersionID, limit, offset)
		if err != nil {
			return nil, err
		}

		var roomsWithStatus []models.RoomWithJoinStatus
		for _, room := range normalRooms {
			roomsWithStatus = append(roomsWithStatus, models.RoomWithJoinStatus{
				Room:     room,
				IsJoined: false,
			})
		}
		return roomsWithStatus, nil
	}

	// 1つのクエリで部屋一覧とユーザーの参加状態を同時に取得（最適化）
	var roomsWithStatus []models.RoomWithJoinStatus
	query := `
		SELECT 
			rooms.*,
			gv.name as game_version_name,
			gv.code as game_version_code,
			u.display_name as host_display_name,
			u.psn_online_id as host_psn_online_id,
			COUNT(DISTINCT rm_all.id) as current_players,
			CASE WHEN rm_user.id IS NOT NULL THEN true ELSE false END as is_joined
		FROM rooms
		LEFT JOIN game_versions gv ON rooms.game_version_id = gv.id
		LEFT JOIN users u ON rooms.host_user_id = u.id
		LEFT JOIN room_members rm_all ON rooms.id = rm_all.room_id AND rm_all.status = 'active'
		LEFT JOIN room_members rm_user ON rooms.id = rm_user.room_id AND rm_user.user_id = ? AND rm_user.status = 'active'
		WHERE rooms.is_active = true
	`

	params := []interface{}{*userID}

	if gameVersionID != nil {
		query += " AND rooms.game_version_id = ?"
		params = append(params, *gameVersionID)
	}

	query += `
		GROUP BY rooms.id, gv.id, u.id, rm_user.id
		ORDER BY 
			CASE WHEN rm_user.id IS NOT NULL THEN 0 ELSE 1 END,
			rooms.created_at DESC
		LIMIT ? OFFSET ?
	`
	params = append(params, limit, offset)

	type roomQueryResult struct {
		models.Room
		GameVersionName string  `json:"game_version_name"`
		GameVersionCode string  `json:"game_version_code"`
		HostDisplayName string  `json:"host_display_name"`
		HostPSNOnlineID *string `json:"host_psn_online_id"`
		CurrentPlayers  int     `json:"current_players"`
		IsJoined        bool    `json:"is_joined"`
	}

	var results []roomQueryResult
	if err := r.db.GetConn().Raw(query, params...).Scan(&results).Error; err != nil {
		return nil, err
	}

	// 結果をRoomWithJoinStatusに変換
	for _, result := range results {
		// GameVersionとHostの情報を設定
		result.Room.GameVersion = models.GameVersion{
			ID:   result.Room.GameVersionID,
			Name: result.GameVersionName,
			Code: result.GameVersionCode,
		}
		result.Room.Host = models.User{
			ID:          result.Room.HostUserID,
			DisplayName: result.HostDisplayName,
			PSNOnlineID: result.HostPSNOnlineID,
		}
		result.Room.CurrentPlayers = result.CurrentPlayers

		roomsWithStatus = append(roomsWithStatus, models.RoomWithJoinStatus{
			Room:     result.Room,
			IsJoined: result.IsJoined,
		})
	}

	return roomsWithStatus, nil
}

func (r *roomRepository) UpdateRoom(room *models.Room) error {
	return r.db.GetConn().Save(room).Error
}

func (r *roomRepository) ToggleRoomClosed(id uuid.UUID, isClosed bool) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Update("is_closed", isClosed).Error
}

func (r *roomRepository) IncrementRoomPlayerCount(id uuid.UUID) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Update("current_players", gorm.Expr("current_players + ?", 1)).Error
}

func (r *roomRepository) DecrementRoomPlayerCount(id uuid.UUID) error {
	return r.db.GetConn().
		Model(&models.Room{}).
		Where("id = ?", id).
		Where("current_players > ?", 0).
		Update("current_players", gorm.Expr("current_players - ?", 1)).Error
}

func (r *roomRepository) JoinRoom(roomID, userID uuid.UUID, password string) error {
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		// ユーザーの存在確認（開発環境では自動作成）
		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 開発環境でのみダミーユーザーを作成
				user = models.User{
					ID:             userID,
					SupabaseUserID: userID, // 開発用
					Email:          fmt.Sprintf("dev-user-%s@example.com", userID.String()[:8]),
					DisplayName:    fmt.Sprintf("開発ユーザー_%s", userID.String()[:8]),
					IsActive:       true,
					Role:           "user",
				}
				if createErr := tx.Create(&user).Error; createErr != nil {
					return fmt.Errorf("開発用ユーザー作成に失敗しました: %w", createErr)
				}
			} else {
				return err
			}
		}

		// 他の部屋に参加しているかチェック
		var activeInOtherRoom models.RoomMember
		query := "user_id = ? AND status = ? AND room_id != ?"
		if err := tx.Where(query, userID, "active", roomID).First(&activeInOtherRoom).Error; err == nil {
			// 既に他の部屋に参加している場合
			return fmt.Errorf("OTHER_ROOM_ACTIVE:既に別の部屋に参加しています")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// record not found 以外のエラーが発生した場合
			return fmt.Errorf("部屋メンバー検索エラー: %w", err)
		}

		var room models.Room
		if err := tx.Where("id = ?", roomID).First(&room).Error; err != nil {
			return err
		}

		if !room.CanJoin() {
			return fmt.Errorf("ルームに参加できません")
		}

		if !room.CheckPassword(password) {
			return fmt.Errorf("パスワードが間違っています")
		}

		// アクティブなメンバーかチェック
		var activeMember models.RoomMember
		if err := tx.Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "active").
			First(&activeMember).Error; err == nil {
			// 既に参加している場合は特別なエラーコードを返す
			return fmt.Errorf("ALREADY_JOINED:既にルームに参加しています")
		}

		// 既存の退室済みメンバーがいるかチェック
		var leftMember models.RoomMember
		if err := tx.Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "left").
			First(&leftMember).Error; err == nil {
			// 既存の退室済みレコードを再アクティブ化
			if err := tx.Model(&leftMember).Updates(map[string]interface{}{
				"status":    "active",
				"joined_at": time.Now(),
				"left_at":   nil,
			}).Error; err != nil {
				return err
			}
		} else {
			// 新規メンバー作成
			var maxPlayerNumber int
			tx.Model(&models.RoomMember{}).
				Where("room_id = ? AND status = ?", roomID, "active").
				Select("COALESCE(MAX(player_number), 0)").
				Scan(&maxPlayerNumber)

			member := models.RoomMember{
				RoomID:       roomID,
				UserID:       userID,
				PlayerNumber: maxPlayerNumber + 1,
				Status:       "active",
				JoinedAt:     time.Now(),
			}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}

		// 部屋のプレイヤー数を更新
		if err := tx.Model(&models.Room{}).
			Where("id = ?", roomID).
			Update("current_players", gorm.Expr("current_players + ?", 1)).Error; err != nil {
			return err
		}

		// 入室ログを記録
		log := models.RoomLog{
			RoomID: roomID,
			UserID: &userID,
			Action: "join",
			Details: models.JSONB{
				"user_name": user.DisplayName,
			},
		}
		return tx.Create(&log).Error
	})
}

func (r *roomRepository) LeaveRoom(roomID, userID uuid.UUID) error {
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&models.RoomMember{}).
			Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, "active").
			Updates(map[string]interface{}{
				"status":  "left",
				"left_at": &now,
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("ルームメンバーが見つかりません")
		}

		// ユーザー情報を取得
		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}

		// 部屋のプレイヤー数を更新
		if err := tx.Model(&models.Room{}).
			Where("id = ?", roomID).
			Where("current_players > ?", 0).
			Update("current_players", gorm.Expr("current_players - ?", 1)).Error; err != nil {
			return err
		}

		// 退室ログを記録
		log := models.RoomLog{
			RoomID: roomID,
			UserID: &userID,
			Action: "leave",
			Details: models.JSONB{
				"user_name": user.DisplayName,
			},
		}
		return tx.Create(&log).Error
	})
}

func (r *roomRepository) FindActiveRoomByUserID(userID uuid.UUID) (*models.Room, error) {
	var member models.RoomMember
	err := r.db.GetConn().
		Where("user_id = ? AND status = ?", userID, "active").
		First(&member).Error
	if err != nil {
		return nil, err
	}

	var room models.Room
	err = r.db.GetConn().
		Where("id = ?", member.RoomID).
		First(&room).Error
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *roomRepository) IsUserJoinedRoom(roomID, userID uuid.UUID) bool {
	var member models.RoomMember
	err := r.db.GetConn().Where("room_id = ? AND user_id = ? AND status = ?",
		roomID, userID, "active").First(&member).Error
	return err == nil
}

func (r *roomRepository) GetRoomMembers(roomID uuid.UUID) ([]models.RoomMember, error) {
	var members []models.RoomMember
	err := r.db.GetConn().
		Preload("User").
		Where("room_id = ? AND status = ?", roomID, "active").
		Order("is_host DESC, joined_at ASC").
		Find(&members).Error

	if err != nil {
		return nil, err
	}

	// ホストユーザーの確認と設定
	var room models.Room
	if err := r.db.GetConn().Where("id = ?", roomID).First(&room).Error; err == nil {
		for i := range members {
			members[i].IsHost = members[i].UserID == room.HostUserID
		}
	}

	return members, nil
}

func (r *roomRepository) GetRoomLogs(roomID uuid.UUID) ([]models.RoomLog, error) {
	var logs []models.RoomLog
	err := r.db.GetConn().
		Preload("User").
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&logs).Error

	return logs, err
}
