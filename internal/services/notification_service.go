package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

// NotificationService ユーザー宛のお知らせを作成するサービス
type NotificationService struct {
	repo *repository.Repository
}

// NewNotificationService 新しいNotificationServiceインスタンスを作成
func NewNotificationService(repo *repository.Repository) *NotificationService {
	return &NotificationService{repo: repo}
}

// NotifyRoomAutoDismissed 作成した部屋が自動削除されたことをホストに知らせる
func (s *NotificationService) NotifyRoomAutoDismissed(room *models.Room) error {
	if room == nil {
		return fmt.Errorf("room is nil")
	}

	return s.repo.Notification.Create(&models.Notification{
		UserID:  room.HostUserID,
		Type:    models.NotificationRoomAutoDismissed,
		Title:   fmt.Sprintf("部屋「%s」が自動的に削除されました", room.Name),
		Body:    stringPtr("一定期間利用がなかったため、部屋を自動的に削除しました。必要であれば新しく部屋を作成してください。"),
		LinkURL: stringPtr("/profile"),
	})
}

// NotifyRoomKicked 部屋から退出させられたことを本人に知らせる
func (s *NotificationService) NotifyRoomKicked(userID uuid.UUID, room *models.Room) error {
	if userID == uuid.Nil || room == nil {
		return fmt.Errorf("invalid input: userID=%v room=%v", userID, room)
	}

	return s.repo.Notification.Create(&models.Notification{
		UserID:      userID,
		Type:        models.NotificationRoomKicked,
		Title:       fmt.Sprintf("部屋「%s」から退出となりました", room.Name),
		Body:        stringPtr("ホストにより部屋から退出させられました。この部屋には再度参加できません。"),
		LinkURL:     stringPtr("/rooms"),
		ActorUserID: &room.HostUserID,
	})
}

// NotifyFollowed フォローされたことを本人に知らせる
func (s *NotificationService) NotifyFollowed(followerID, followingID uuid.UUID, follower *models.User) error {
	if followerID == uuid.Nil || followingID == uuid.Nil || follower == nil {
		return fmt.Errorf("invalid input: followerID=%v followingID=%v", followerID, followingID)
	}

	name := follower.DisplayName
	if name == "" && follower.Username != nil {
		name = *follower.Username
	}

	return s.repo.Notification.Create(&models.Notification{
		UserID:      followingID,
		Type:        models.NotificationFollow,
		Title:       fmt.Sprintf("%sさんにフォローされました", name),
		LinkURL:     stringPtr("/users/" + followerID.String()),
		ActorUserID: &followerID,
	})
}

// NotifyRoomDismissedToMembers ホストが部屋を解散したことを参加中のメンバー（ホスト以外）に知らせる
func (s *NotificationService) NotifyRoomDismissedToMembers(room *models.Room, members []models.RoomMember) error {
	if room == nil {
		return fmt.Errorf("room is nil")
	}

	return s.notifyMembers(room, members, models.NotificationRoomDismissed,
		fmt.Sprintf("部屋「%s」が解散されました", room.Name),
		"ホストにより部屋が解散されました。別の部屋を探すか、新しく部屋を作成してください。")
}

// NotifyRoomAutoDismissedToMembers 参加していた部屋が自動削除されたことをメンバー（ホスト以外）に知らせる
func (s *NotificationService) NotifyRoomAutoDismissedToMembers(room *models.Room, members []models.RoomMember) error {
	if room == nil {
		return fmt.Errorf("room is nil")
	}

	return s.notifyMembers(room, members, models.NotificationRoomAutoDismissed,
		fmt.Sprintf("参加していた部屋「%s」が自動的に削除されました", room.Name),
		"一定期間利用がなかったため、部屋は自動的に削除されました。")
}

// notifyMembers ホスト以外のメンバー全員に同じ内容のお知らせを作成する。一部失敗しても続行し、まとめて返す
func (s *NotificationService) notifyMembers(room *models.Room, members []models.RoomMember, notificationType, title, body string) error {
	var errs []error
	for _, member := range members {
		if member.UserID == uuid.Nil || member.UserID == room.HostUserID {
			continue
		}
		err := s.repo.Notification.Create(&models.Notification{
			UserID:      member.UserID,
			Type:        notificationType,
			Title:       title,
			Body:        stringPtr(body),
			LinkURL:     stringPtr("/rooms"),
			ActorUserID: &room.HostUserID,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("notify member %s: %w", member.UserID, err))
		}
	}

	return errors.Join(errs...)
}
