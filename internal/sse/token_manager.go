package sse

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenInvalid  = errors.New("token invalid")
)

var (
	globalTokenManager TokenManager
	tokenManagerMu     sync.RWMutex
)

type TokenInfo struct {
	RoomID    uuid.UUID `json:"room_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenManager interface {
	GenerateToken(roomID, userID uuid.UUID) (string, error)
	ConsumeToken(token string) (*TokenInfo, error)
}

func GetTokenManager() TokenManager {
	tokenManagerMu.RLock()
	defer tokenManagerMu.RUnlock()

	if globalTokenManager == nil {
		// 未初期化の場合のフォールバック
		globalTokenManager = NewInMemoryTokenManager(5 * time.Minute)
	}

	return globalTokenManager
}

func SetTokenManager(manager TokenManager) {
	tokenManagerMu.Lock()
	defer tokenManagerMu.Unlock()

	globalTokenManager = manager
}

func generateSecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// 暗号学的に安全な乱数生成に失敗した場合のフォールバック
		return uuid.New().String()
	}

	return hex.EncodeToString(b)
}
