package models

// AllModels AutoMigrate 対象の全モデル。
// マイグレーションのリストはここだけで管理する（turso / postgres で共用）
func AllModels() []interface{} {
	return []interface{}{
		&Platform{},
		&GameVersion{},
		&User{},
		&Room{},
		&RoomMember{},
		&RoomMessage{},
		&MessageReaction{},
		&ReactionType{},
		&UserBlock{},
		&PlayerName{},
		&UserFollow{},
		&UserActivity{},
		&RoomLog{},
		&PasswordReset{},
		&UserReport{},
		&ReportAttachment{},
		&Contact{},
		&Notification{},
		&UserNotificationState{},
	}
}
