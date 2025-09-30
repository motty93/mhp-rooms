package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
)

type roomRepository struct {
	db DBInterface
}

func NewRoomRepository(db DBInterface) RoomRepository {
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
				Data: map[string]interface{}{
					"room_name": room.Name,
				},
			},
		}
		return tx.Create(&log).Error
	})
}

func (r *roomRepository) FindRoomByID(id uuid.UUID) (*models.Room, error) {
	var room models.Room
	err := r.db.GetConn().
		Preload("GameVersion").
		Preload("Host", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "is_active", "role", "created_at", "updated_at")
		}).
		Where("id = ?", id).
		First(&room).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
		Preload("Host", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "is_active", "role", "created_at", "updated_at")
		}).
		Where("room_code = ?", roomCode).
		First(&room).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ルームが見つかりません")
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) RoomCodeExists(roomCode string) (bool, error) {
	var count int64
	err := r.db.GetConn().Model(&models.Room{}).
		Where("room_code = ?", roomCode).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *roomRepository) GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error) {
	// 効率的なクエリ: 1つのクエリで部屋情報と参加者数を取得
	type roomQueryResult struct {
		models.Room
		GameVersionName string  `json:"game_version_name"`
		GameVersionCode string  `json:"game_version_code"`
		HostDisplayName string  `json:"host_display_name"`
		HostPSNOnlineID *string `json:"host_psn_online_id"`
		CurrentPlayers  int     `json:"current_players"`
	}

	var results []roomQueryResult
	sqlQuery := `
		SELECT
			rooms.id, rooms.room_code, rooms.name, rooms.description,
			rooms.game_version_id, rooms.host_user_id, rooms.max_players,
			rooms.password_hash, rooms.target_monster, rooms.rank_requirement,
			rooms.is_active, rooms.is_closed, rooms.created_at, rooms.updated_at, rooms.closed_at,
			gv.name as game_version_name,
			gv.code as game_version_code,
			u.display_name as host_display_name,
			u.psn_online_id as host_psn_online_id,
			COUNT(DISTINCT rm.id) as current_players
		FROM rooms
		LEFT JOIN game_versions gv ON rooms.game_version_id = gv.id
		LEFT JOIN users u ON rooms.host_user_id = u.id
		LEFT JOIN room_members rm ON rooms.id = rm.room_id AND rm.status = 'active'
		WHERE rooms.is_active = true`

	params := []interface{}{}
	if gameVersionID != nil {
		sqlQuery += " AND rooms.game_version_id = ?"
		params = append(params, *gameVersionID)
	}

	sqlQuery += `
		GROUP BY rooms.id, gv.id, u.id
		ORDER BY rooms.created_at DESC
		LIMIT ? OFFSET ?`
	params = append(params, limit, offset)

	if err := r.db.GetConn().Raw(sqlQuery, params...).Scan(&results).Error; err != nil {
		return nil, err
	}

	// 結果をRoomに変換
	rooms := make([]models.Room, len(results))
	for i, result := range results {
		rooms[i] = result.Room
		rooms[i].GameVersion = models.GameVersion{
			BaseModel: models.BaseModel{
				ID: result.Room.GameVersionID,
			},
			Name: result.GameVersionName,
			Code: result.GameVersionCode,
		}
		rooms[i].Host = models.User{
			BaseModel: models.BaseModel{
				ID: result.Room.HostUserID,
			},
			DisplayName: result.HostDisplayName,
			PSNOnlineID: result.HostPSNOnlineID,
		}
		rooms[i].CurrentPlayers = result.CurrentPlayers
	}

	return rooms, nil
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

	// 効率的なクエリ: サブクエリを使用してroom_membersの重複JOINを回避
	var roomsWithStatus []models.RoomWithJoinStatus
	query := `
		SELECT
			rooms.id, rooms.room_code, rooms.name, rooms.description,
			rooms.game_version_id, rooms.host_user_id, rooms.max_players,
			rooms.password_hash, rooms.target_monster, rooms.rank_requirement,
			rooms.is_active, rooms.is_closed, rooms.created_at, rooms.updated_at, rooms.closed_at,
			gv.name as game_version_name,
			gv.code as game_version_code,
			u.display_name as host_display_name,
			u.psn_online_id as host_psn_online_id,
			COALESCE(member_counts.current_players, 0) as current_players,
			COALESCE(user_membership.is_joined, false) as is_joined
		FROM rooms
		LEFT JOIN game_versions gv ON rooms.game_version_id = gv.id
		LEFT JOIN users u ON rooms.host_user_id = u.id
		LEFT JOIN (
			SELECT room_id, COUNT(*) as current_players
			FROM room_members
			WHERE status = 'active'
			GROUP BY room_id
		) member_counts ON rooms.id = member_counts.room_id
		LEFT JOIN (
			SELECT room_id, true as is_joined
			FROM room_members
			WHERE user_id = ? AND status = 'active'
		) user_membership ON rooms.id = user_membership.room_id
		WHERE rooms.is_active = true
	`

	params := []interface{}{*userID}

	if gameVersionID != nil {
		query += " AND rooms.game_version_id = ?"
		params = append(params, *gameVersionID)
	}

	query += `
		ORDER BY
			CASE WHEN user_membership.is_joined IS NOT NULL THEN 0 ELSE 1 END,
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
			BaseModel: models.BaseModel{
				ID: result.Room.GameVersionID,
			},
			Name: result.GameVersionName,
			Code: result.GameVersionCode,
		}
		result.Room.Host = models.User{
			BaseModel: models.BaseModel{
				ID: result.Room.HostUserID,
			},
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
		if err := tx.Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "is_active", "role", "created_at", "updated_at").Where("id = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 開発環境でのみダミーユーザーを作成
				user = models.User{
					BaseModel: models.BaseModel{
						ID: userID,
					},
					SupabaseUserID: userID, // 開発用
					Email:          fmt.Sprintf("dev-user-%s@example.com", userID.String()[:8]),
					DisplayName:    fmt.Sprintf("開発ユーザー_%s", userID.String()[:8]),
					IsActive:       true,
					Role:           "dummy",
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
				Data: map[string]interface{}{
					"user_name": user.DisplayName,
				},
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
		if err := tx.Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "is_active", "role", "created_at", "updated_at").Where("id = ?", userID).First(&user).Error; err != nil {
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
				Data: map[string]interface{}{
					"user_name": user.DisplayName,
				},
			},
		}
		return tx.Create(&log).Error
	})
}

func (r *roomRepository) FindActiveRoomByUserID(userID uuid.UUID) (*models.Room, error) {
	var member models.RoomMember

	// Limit(1)を使用してrecord not foundエラーのログを回避
	result := r.db.GetConn().
		Where("user_id = ? AND status = ?", userID, "active").
		Limit(1).
		Find(&member)

	// レコードが見つからない場合（RowsAffectedが0の場合）
	if result.RowsAffected == 0 {
		return nil, nil
	}

	// その他のエラーチェック
	if result.Error != nil {
		return nil, result.Error
	}

	var room models.Room
	err := r.db.GetConn().
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
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "is_active", "role", "created_at", "updated_at")
		}).
		Where("room_id = ? AND status = ?", roomID, "active").
		Order("is_host DESC, joined_at ASC").
		Find(&members).Error

	if err != nil {
		return nil, err
	}

	return members, nil
}

func (r *roomRepository) GetRoomLogs(roomID uuid.UUID) ([]models.RoomLog, error) {
	var logs []models.RoomLog
	err := r.db.GetConn().
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "supabase_user_id", "email", "username", "display_name", "avatar_url", "bio", "psn_online_id", "nintendo_network_id", "nintendo_switch_id", "pretendo_network_id", "twitter_id", "is_active", "role", "created_at", "updated_at")
		}).
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&logs).Error

	return logs, err
}

// UpdateRoom 部屋情報を更新
func (r *roomRepository) UpdateRoom(room *models.Room) error {
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		// 部屋情報を更新
		if err := tx.Save(room).Error; err != nil {
			return err
		}

		// 更新ログを記録
		log := models.RoomLog{
			RoomID: room.ID,
			UserID: &room.HostUserID,
			Action: "update_settings",
			Details: models.JSONB{
				Data: map[string]interface{}{
					"room_name":            room.Name,
					"max_players":          room.MaxPlayers,
					"game_version_id":      room.GameVersionID,
					"has_description":      room.Description != nil,
					"has_target_monster":   room.TargetMonster != nil,
					"has_rank_requirement": room.RankRequirement != nil,
					"has_password":         room.PasswordHash != nil,
				},
			},
		}
		return tx.Create(&log).Error
	})
}

// DismissRoom 部屋を解散
func (r *roomRepository) DismissRoom(roomID uuid.UUID) error {
	return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
		// 部屋情報を取得
		var room models.Room
		if err := tx.First(&room, "id = ?", roomID).Error; err != nil {
			return err
		}

		// 全メンバーを退出状態に変更
		if err := tx.Model(&models.RoomMember{}).
			Where("room_id = ? AND status = ?", roomID, "active").
			Updates(map[string]interface{}{
				"status":  "left",
				"left_at": time.Now(),
			}).Error; err != nil {
			return err
		}

		// 部屋を非アクティブに変更
		if err := tx.Model(&room).Updates(map[string]interface{}{
			"is_active":       false,
			"current_players": 0,
			"updated_at":      time.Now(),
		}).Error; err != nil {
			return err
		}

		// 解散ログを記録
		log := models.RoomLog{
			RoomID: roomID,
			UserID: &room.HostUserID,
			Action: "dismiss",
			Details: models.JSONB{
				Data: map[string]interface{}{
					"room_name": room.Name,
				},
			},
		}
		return tx.Create(&log).Error
	})
}

// GetUserRoomStatus ユーザーの部屋状態を取得
func (r *roomRepository) GetUserRoomStatus(userID uuid.UUID) (string, *models.Room, error) {
	// 1. ホストとして部屋を持っているかチェック
	var hostRooms []models.Room
	result := r.db.GetConn().
		Preload("GameVersion").
		Where("host_user_id = ? AND is_active = ? AND is_closed = ?", userID, true, false).
		Limit(1).
		Find(&hostRooms)

	if result.Error != nil {
		return "", nil, result.Error
	}

	if len(hostRooms) > 0 {
		// ホストとして部屋を持っている
		return "HOST", &hostRooms[0], nil
	}

	// 2. 参加者として部屋に参加しているかチェック
	var members []models.RoomMember
	result = r.db.GetConn().
		Where("user_id = ? AND status = ? AND is_host = ?", userID, "active", false).
		Limit(1).
		Find(&members)

	if result.Error != nil {
		return "", nil, result.Error
	}

	if len(members) > 0 {
		// 参加者として部屋に参加している
		var guestRoom models.Room
		err := r.db.GetConn().
			Preload("GameVersion").
			Where("id = ?", members[0].RoomID).
			First(&guestRoom).Error
		if err != nil {
			return "", nil, err
		}
		return "GUEST", &guestRoom, nil
	}

	// 3. どの部屋にも所属していない
	return "NONE", nil, nil
}

// GetRoomsByHostUser ホストユーザーが作成した部屋一覧を取得
func (r *roomRepository) GetRoomsByHostUser(userID uuid.UUID, limit, offset int) ([]models.Room, error) {
	// 最適化されたクエリ: JOINを使用してN+1問題を解決
	var results []struct {
		models.Room
		GameVersionName string  `json:"game_version_name"`
		GameVersionCode string  `json:"game_version_code"`
		HostDisplayName string  `json:"host_display_name"`
		HostPSNOnlineID *string `json:"host_psn_online_id"`
		CurrentPlayers  int     `json:"current_players"`
	}

	query := `
		SELECT
			rooms.*,
			gv.name as game_version_name,
			gv.code as game_version_code,
			u.display_name as host_display_name,
			u.psn_online_id as host_psn_online_id,
			COUNT(DISTINCT rm.id) as current_players
		FROM rooms
		LEFT JOIN game_versions gv ON rooms.game_version_id = gv.id
		LEFT JOIN users u ON rooms.host_user_id = u.id
		LEFT JOIN room_members rm ON rooms.id = rm.room_id AND rm.status = 'active'
		WHERE rooms.host_user_id = ?
		GROUP BY rooms.id, gv.id, u.id
		ORDER BY rooms.created_at DESC
		LIMIT ? OFFSET ?
	`

	if err := r.db.GetConn().Raw(query, userID, limit, offset).Scan(&results).Error; err != nil {
		return nil, err
	}

	// 結果をmodels.Roomに変換
	var rooms []models.Room
	for _, result := range results {
		// GameVersionとHostの情報を設定
		result.Room.GameVersion = models.GameVersion{
			BaseModel: models.BaseModel{
				ID: result.Room.GameVersionID,
			},
			Name: result.GameVersionName,
			Code: result.GameVersionCode,
		}
		result.Room.Host = models.User{
			BaseModel: models.BaseModel{
				ID: result.Room.HostUserID,
			},
			DisplayName: result.HostDisplayName,
			PSNOnlineID: result.HostPSNOnlineID,
		}
		result.Room.CurrentPlayers = result.CurrentPlayers

		rooms = append(rooms, result.Room)
	}

	return rooms, nil
}
