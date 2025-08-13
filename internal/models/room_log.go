package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// JSONB はPostgreSQLのJSONBフィールド用のカスタム型
type JSONB struct {
	Data interface{}
}

// Value はdriver.Valuerインターフェースを実装
func (j JSONB) Value() (driver.Value, error) {
	if j.Data == nil {
		return nil, nil
	}

	return json.Marshal(j.Data)
}

// Scan はsql.Scannerインターフェースを実装
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		j.Data = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	// まず汎用的なinterface{}にunmarshal
	var result interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	
	j.Data = result
	return nil
}

// RoomLog はルームアクションの監査ログ
type RoomLog struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoomID    uuid.UUID  `gorm:"type:uuid;not null" json:"room_id"`
	UserID    *uuid.UUID `gorm:"type:uuid" json:"user_id"`
	Action    string     `gorm:"type:varchar(50);not null" json:"action"`
	Details   JSONB      `gorm:"type:jsonb" json:"details"`
	CreatedAt time.Time  `json:"created_at"`

	// リレーション
	Room Room  `gorm:"foreignKey:RoomID" json:"room"`
	User *User `gorm:"foreignKey:UserID" json:"user"`
}
