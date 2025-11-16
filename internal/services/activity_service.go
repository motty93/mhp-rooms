package services

import (
	"fmt"
	"log"

	"github.com/google/uuid"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

// ActivityService はアクティビティ記録・管理を行うサービス
type ActivityService struct {
	repo *repository.Repository
}

// NewActivityService は新しいActivityServiceインスタンスを作成
func NewActivityService(repo *repository.Repository) *ActivityService {
	return &ActivityService{repo: repo}
}

// RecordRoomCreate 部屋作成のアクティビティを記録
func (s *ActivityService) RecordRoomCreate(userID uuid.UUID, room *models.Room) error {
	if userID == uuid.Nil || room == nil {
		return fmt.Errorf("無効な入力: userID=%v, room=%v", userID, room)
	}

	// ゲームバージョンの情報を取得
	gameVersion, err := s.repo.GameVersion.FindGameVersionByID(room.GameVersionID)
	if err != nil {
		log.Printf("ゲームバージョン取得エラー: %v", err)
		// エラーでも処理を続行（アクティビティ記録が主処理を止めるべきではない）
	}

	gameVersionCode := "不明"
	if gameVersion != nil {
		gameVersionCode = gameVersion.Code
	}

	// アクティビティメタデータを構築
	metadata := models.RoomActivityMetadata{
		GameVersion:   gameVersionCode,
		MaxPlayers:    room.MaxPlayers,
		TargetMonster: getStringValue(room.TargetMonster),
		HostUserID:    room.HostUserID.String(),
		RoomPassword:  room.PasswordHash != nil && *room.PasswordHash != "",
	}

	activity := &models.UserActivity{
		UserID:            userID,
		ActivityType:      models.ActivityRoomCreate,
		Title:             fmt.Sprintf("【部屋作成】%s", room.Name),
		Description:       buildRoomDescription(room),
		RelatedEntityType: stringPtr(models.EntityTypeRoom),
		RelatedEntityID:   &room.ID,
		Icon:              "fa-door-open",
		IconColor:         "text-green-500",
	}

	if err := activity.SetMetadata(metadata); err != nil {
		log.Printf("メタデータ設定エラー: %v", err)
		// メタデータ設定に失敗してもアクティビティ記録は続行
	}

	return s.repo.UserActivity.CreateActivity(activity)
}

// RecordRoomJoin 部屋参加のアクティビティを記録
func (s *ActivityService) RecordRoomJoin(userID uuid.UUID, room *models.Room, hostUser *models.User) error {
	if userID == uuid.Nil || room == nil || hostUser == nil {
		return fmt.Errorf("無効な入力: userID=%v, room=%v, hostUser=%v", userID, room, hostUser)
	}

	// ゲームバージョンの情報を取得
	gameVersion, err := s.repo.GameVersion.FindGameVersionByID(room.GameVersionID)
	if err != nil {
		log.Printf("ゲームバージョン取得エラー: %v", err)
	}

	gameVersionCode := "不明"
	if gameVersion != nil {
		gameVersionCode = gameVersion.Code
	}

	// アクティビティメタデータを構築
	metadata := models.RoomActivityMetadata{
		GameVersion: gameVersionCode,
		HostUserID:  room.HostUserID.String(),
	}

	// DisplayNameが空の場合はUsernameを使用
	hostDisplayName := hostUser.DisplayName
	if hostDisplayName == "" && hostUser.Username != nil {
		hostDisplayName = *hostUser.Username
	}

	activity := &models.UserActivity{
		UserID:            userID,
		ActivityType:      models.ActivityRoomJoin,
		Title:             fmt.Sprintf("【部屋参加】%s", room.Name),
		Description:       stringPtr(fmt.Sprintf("ホスト: %s", hostDisplayName)),
		RelatedEntityType: stringPtr(models.EntityTypeRoom),
		RelatedEntityID:   &room.ID,
		Icon:              "fa-right-to-bracket",
		IconColor:         "text-blue-500",
	}

	if err := activity.SetMetadata(metadata); err != nil {
		log.Printf("メタデータ設定エラー: %v", err)
	}

	return s.repo.UserActivity.CreateActivity(activity)
}

// RecordRoomLeave 部屋退出のアクティビティを記録
func (s *ActivityService) RecordRoomLeave(userID uuid.UUID, room *models.Room) error {
	if userID == uuid.Nil || room == nil {
		return fmt.Errorf("無効な入力: userID=%v, room=%v", userID, room)
	}

	// ゲームバージョンの情報を取得
	gameVersion, err := s.repo.GameVersion.FindGameVersionByID(room.GameVersionID)
	if err != nil {
		log.Printf("ゲームバージョン取得エラー: %v", err)
	}

	gameVersionCode := "不明"
	if gameVersion != nil {
		gameVersionCode = gameVersion.Code
	}

	// アクティビティメタデータを構築
	metadata := models.RoomActivityMetadata{
		GameVersion: gameVersionCode,
		HostUserID:  room.HostUserID.String(),
	}

	activity := &models.UserActivity{
		UserID:            userID,
		ActivityType:      models.ActivityRoomLeave,
		Title:             fmt.Sprintf("【部屋退出】%s", room.Name),
		Description:       nil,
		RelatedEntityType: stringPtr(models.EntityTypeRoom),
		RelatedEntityID:   &room.ID,
		Icon:              "fa-right-from-bracket",
		IconColor:         "text-red-500",
	}

	if err := activity.SetMetadata(metadata); err != nil {
		log.Printf("メタデータ設定エラー: %v", err)
	}

	return s.repo.UserActivity.CreateActivity(activity)
}

// RecordRoomClose 部屋終了のアクティビティを記録
func (s *ActivityService) RecordRoomClose(userID uuid.UUID, room *models.Room) error {
	if userID == uuid.Nil || room == nil {
		return fmt.Errorf("無効な入力: userID=%v, room=%v", userID, room)
	}

	activity := &models.UserActivity{
		UserID:            userID,
		ActivityType:      models.ActivityRoomClose,
		Title:             fmt.Sprintf("【部屋終了】%s", room.Name),
		Description:       stringPtr("部屋を終了しました"),
		RelatedEntityType: stringPtr(models.EntityTypeRoom),
		RelatedEntityID:   &room.ID,
		Icon:              "fa-door-closed",
		IconColor:         "text-gray-500",
	}

	return s.repo.UserActivity.CreateActivity(activity)
}

// RecordFollow フォロー開始のアクティビティを記録
func (s *ActivityService) RecordFollow(followerID, followingID uuid.UUID, followingUser *models.User) error {
	if followerID == uuid.Nil || followingID == uuid.Nil || followingUser == nil {
		return fmt.Errorf("無効な入力: followerID=%v, followingID=%v, followingUser=%v", followerID, followingID, followingUser)
	}

	// 相互フォローかチェック
	isMutual, err := s.repo.UserFollow.IsMutualFollow(followerID, followingID)
	if err != nil {
		log.Printf("相互フォローチェックエラー: %v", err)
		isMutual = false // エラー時はfalseで続行
	}

	// アクティビティメタデータを構築
	metadata := models.FollowActivityMetadata{
		FollowingUserID: followingID.String(),
		IsMutualFollow:  isMutual,
	}

	// DisplayNameが空の場合はUsernameを使用
	displayName := followingUser.DisplayName
	if displayName == "" && followingUser.Username != nil {
		displayName = *followingUser.Username
	}

	title := fmt.Sprintf("%sさんをフォローしました", displayName)
	if isMutual {
		title = fmt.Sprintf("%sさんと相互フォローになりました", displayName)
	}

	activity := &models.UserActivity{
		UserID:            followerID,
		ActivityType:      models.ActivityFollowAdd,
		Title:             title,
		Description:       nil,
		RelatedEntityType: stringPtr(models.EntityTypeUser),
		RelatedEntityID:   &followingID,
		Icon:              "fa-user-plus",
		IconColor:         "text-yellow-500",
	}

	if err := activity.SetMetadata(metadata); err != nil {
		log.Printf("メタデータ設定エラー: %v", err)
	}

	return s.repo.UserActivity.CreateActivity(activity)
}

// RecordUserJoin ユーザー登録のアクティビティを記録
func (s *ActivityService) RecordUserJoin(userID uuid.UUID, registrationMethod string) error {
	if userID == uuid.Nil {
		return fmt.Errorf("無効な入力: userID=%v", userID)
	}

	// アクティビティメタデータを構築
	metadata := models.UserJoinActivityMetadata{
		RegistrationMethod: registrationMethod,
	}

	activity := &models.UserActivity{
		UserID:            userID,
		ActivityType:      models.ActivityUserJoin,
		Title:             "HuntersHubに参加しました",
		Description:       stringPtr("新しいハンターとしてコミュニティに参加しました"),
		RelatedEntityType: stringPtr(models.EntityTypeUser),
		RelatedEntityID:   &userID,
		Icon:              "fa-user-check",
		IconColor:         "text-green-500",
	}

	if err := activity.SetMetadata(metadata); err != nil {
		log.Printf("メタデータ設定エラー: %v", err)
	}

	return s.repo.UserActivity.CreateActivity(activity)
}

// ヘルパー関数群

// buildRoomDescription 部屋の詳細説明を構築
func buildRoomDescription(room *models.Room) *string {
	if room.TargetMonster != nil && *room.TargetMonster != "" {
		return stringPtr(fmt.Sprintf("ターゲット: %s", *room.TargetMonster))
	}
	if room.Description != nil && *room.Description != "" {
		return room.Description
	}
	return nil
}

// getStringValue ポインタ文字列から値を安全に取得
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// stringPtr 文字列のポインタを作成
func stringPtr(s string) *string {
	return &s
}
