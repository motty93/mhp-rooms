package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

// RoomCleanupService 一定期間活動がない部屋を自動解散するサービス
type RoomCleanupService struct {
	repo                *repository.Repository
	activityService     *ActivityService
	notificationService *NotificationService
}

// NewRoomCleanupService 新しいRoomCleanupServiceインスタンスを作成
func NewRoomCleanupService(repo *repository.Repository) *RoomCleanupService {
	return &RoomCleanupService{
		repo:                repo,
		activityService:     NewActivityService(repo),
		notificationService: NewNotificationService(repo),
	}
}

// FindInactiveRooms idleDuration の間、活動（作成・設定変更・参加・退出・チャット）がない募集中の部屋を返す
func (s *RoomCleanupService) FindInactiveRooms(idleDuration time.Duration) ([]models.Room, error) {
	if idleDuration <= 0 {
		return nil, fmt.Errorf("idle duration must be positive: %s", idleDuration)
	}

	return s.repo.Room.FindInactiveRooms(time.Now().Add(-idleDuration))
}

// DismissInactiveRooms 非アクティブな部屋を自動解散し、解散できた部屋を返す。
// 一部の部屋で失敗しても残りの処理を続け、失敗分をまとめたエラーを返す
func (s *RoomCleanupService) DismissInactiveRooms(idleDuration time.Duration) ([]models.Room, error) {
	rooms, err := s.FindInactiveRooms(idleDuration)
	if err != nil {
		return nil, fmt.Errorf("find inactive rooms: %w", err)
	}

	var dismissed []models.Room
	var errs []error
	for _, room := range rooms {
		// 解散するとメンバーは退出状態になるため、お知らせ用に事前に取得しておく
		members, err := s.repo.Room.GetRoomMembers(room.ID)
		if err != nil {
			log.Printf("解散前のメンバー取得に失敗: room_id=%s: %v", room.ID, err)
			members = nil
		}

		if err := s.repo.Room.DismissRoom(room.ID, models.DismissReasonInactive); err != nil {
			errs = append(errs, fmt.Errorf("dismiss room %s (%s): %w", room.ID, room.Name, err))
			continue
		}
		dismissed = append(dismissed, room)

		// アクティビティ記録の失敗は解散処理の成否に影響させない
		if err := s.activityService.RecordRoomAutoDismiss(room.HostUserID, &room); err != nil {
			log.Printf("部屋自動削除アクティビティの記録に失敗: room_id=%s: %v", room.ID, err)
		}
		if err := s.notificationService.NotifyRoomAutoDismissed(&room); err != nil {
			log.Printf("部屋自動削除のお知らせ作成に失敗: room_id=%s: %v", room.ID, err)
		}
		if err := s.notificationService.NotifyRoomAutoDismissedToMembers(&room, members); err != nil {
			log.Printf("部屋自動削除のメンバー向けお知らせ作成に失敗: room_id=%s: %v", room.ID, err)
		}
	}

	return dismissed, errors.Join(errs...)
}
