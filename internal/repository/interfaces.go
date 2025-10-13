package repository

import (
	"time"

	"mhp-rooms/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	FindUserByID(id uuid.UUID) (*models.User, error)
	FindUsersByIDs(ids []uuid.UUID) ([]models.User, error)
	FindUserBySupabaseUserID(supabaseUserID uuid.UUID) (*models.User, error)
	FindUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	GetActiveUsers(limit, offset int) ([]models.User, error)
}

type GameVersionRepository interface {
	FindGameVersionByID(id uuid.UUID) (*models.GameVersion, error)
	FindGameVersionByCode(code string) (*models.GameVersion, error)
	GetActiveGameVersions() ([]models.GameVersion, error)
}

type PlatformRepository interface {
	GetActivePlatforms() ([]models.Platform, error)
}

type RoomRepository interface {
	CreateRoom(room *models.Room) error
	FindRoomByID(id uuid.UUID) (*models.Room, error)
	FindRoomByRoomCode(roomCode string) (*models.Room, error)
	RoomCodeExists(roomCode string) (bool, error)
	GetActiveRooms(gameVersionID *uuid.UUID, limit, offset int) ([]models.Room, error)
	GetActiveRoomsWithJoinStatus(userID *uuid.UUID, gameVersionID *uuid.UUID, limit, offset int) ([]models.RoomWithJoinStatus, error)
	CountActiveRooms(gameVersionID *uuid.UUID) (int64, error)
	UpdateRoom(room *models.Room) error
	DismissRoom(id uuid.UUID) error
	ToggleRoomClosed(id uuid.UUID, isClosed bool) error
	IncrementRoomPlayerCount(id uuid.UUID) error
	DecrementRoomPlayerCount(id uuid.UUID) error
	JoinRoom(roomID, userID uuid.UUID, password string) error
	LeaveRoom(roomID, userID uuid.UUID) error
	FindActiveRoomByUserID(userID uuid.UUID) (*models.Room, error)
	IsUserJoinedRoom(roomID, userID uuid.UUID) bool
	GetRoomMembers(roomID uuid.UUID) ([]models.RoomMember, error)
	GetRoomLogs(roomID uuid.UUID) ([]models.RoomLog, error)
	GetUserRoomStatus(userID uuid.UUID) (string, *models.Room, error) // (status, room, error)
	GetRoomsByHostUser(userID uuid.UUID, limit, offset int) ([]models.Room, error)
}

type PlayerNameRepository interface {
	CreatePlayerName(playerName *models.PlayerName) error
	UpdatePlayerName(playerName *models.PlayerName) error
	FindPlayerNameByUserAndGame(userID, gameVersionID uuid.UUID) (*models.PlayerName, error)
	FindAllPlayerNamesByUser(userID uuid.UUID) ([]models.PlayerName, error)
	DeletePlayerName(id uuid.UUID) error
	UpsertPlayerName(playerName *models.PlayerName) error
}

type RoomMessageRepository interface {
	CreateMessage(message *models.RoomMessage) error
	GetMessages(roomID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.RoomMessage, error)
	DeleteMessage(id uuid.UUID) error
}

type UserBlockRepository interface {
	CreateBlock(block *models.UserBlock) error
	DeleteBlock(blockerUserID, blockedUserID uuid.UUID) error
	IsBlocked(blockerUserID, blockedUserID uuid.UUID) (bool, error)
	CheckBlockRelationship(userID, targetUserID uuid.UUID) (bool, bool, error) // (isBlockedByTarget, isBlockingTarget, error)
	CheckRoomMemberBlocks(userID, roomID uuid.UUID) ([]models.User, error)     // ブロック関係のあるメンバーリストを返す
	GetBlockedUsers(blockerUserID uuid.UUID) ([]models.User, error)
	GetBlockingUsers(blockedUserID uuid.UUID) ([]models.User, error)
}

type UserFollowRepository interface {
	CreateFollow(follow *models.UserFollow) error
	DeleteFollow(followerUserID, followingUserID uuid.UUID) error
	GetFollow(followerUserID, followingUserID uuid.UUID) (*models.UserFollow, error)
	UpdateFollowStatus(followerUserID, followingUserID uuid.UUID, status string) error
	GetFollowers(userID uuid.UUID) ([]models.UserFollow, error)
	GetFollowing(userID uuid.UUID) ([]models.UserFollow, error)
	GetMutualFriends(userID uuid.UUID) ([]models.User, error)
	GetFriendCount(userID uuid.UUID) (int64, error)
	IsMutualFollow(userID1, userID2 uuid.UUID) (bool, error)
}

type UserActivityRepository interface {
	CreateActivity(activity *models.UserActivity) error
	GetUserActivities(userID uuid.UUID, limit, offset int) ([]models.UserActivity, error)
	GetUserActivitiesByType(userID uuid.UUID, activityType string, limit, offset int) ([]models.UserActivity, error)
	CountUserActivities(userID uuid.UUID) (int64, error)
	DeleteActivity(id uuid.UUID) error
	DeleteOldActivities(olderThan time.Time) error
}

type ReportRepository interface {
	Create(report *models.UserReport) error
	GetByID(id uuid.UUID) (*models.UserReport, error)
	GetByReportedUserID(userID uuid.UUID, limit int) ([]models.UserReport, error)
	GetByReporterUserID(userID uuid.UUID, limit int) ([]models.UserReport, error)
	GetPendingReports(limit int, offset int) ([]models.UserReport, int64, error)
	UpdateStatus(id uuid.UUID, status models.ReportStatus, adminNote *string) error
	CheckDuplicateReport(reporterID, reportedID uuid.UUID) (bool, error)
	AddAttachment(attachment *models.ReportAttachment) error
	GetAttachmentsByReportID(reportID uuid.UUID) ([]models.ReportAttachment, error)
	DeleteAttachment(id uuid.UUID) error
	GetReportStatsByUserID(userID uuid.UUID) (map[string]int64, error)
	SearchReports(params ReportSearchParams) ([]models.UserReport, int64, error)
	BatchUpdateStatus(ids []uuid.UUID, status models.ReportStatus, adminNote *string) error
}
