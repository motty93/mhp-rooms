package sse

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type InMemoryTokenManager struct {
	tokens map[string]*TokenInfo
	mu     sync.RWMutex
	ttl    time.Duration
}

func NewInMemoryTokenManager(ttl time.Duration) *InMemoryTokenManager {
	manager := &InMemoryTokenManager{
		tokens: make(map[string]*TokenInfo),
		ttl:    ttl,
	}

	go manager.cleanupExpired()

	return manager
}

func (m *InMemoryTokenManager) GenerateToken(roomID, userID uuid.UUID) (string, error) {
	token := generateSecureToken()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens[token] = &TokenInfo{
		RoomID:    roomID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	return token, nil
}

func (m *InMemoryTokenManager) ConsumeToken(token string) (*TokenInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.tokens[token]
	if !exists {
		return nil, ErrTokenNotFound
	}

	if time.Since(info.CreatedAt) > m.ttl {
		delete(m.tokens, token)
		return nil, ErrTokenExpired
	}

	// One-time use: トークンは消費後に削除される
	delete(m.tokens, token)

	return info, nil
}

func (m *InMemoryTokenManager) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for token, info := range m.tokens {
			if now.Sub(info.CreatedAt) > m.ttl {
				delete(m.tokens, token)
			}
		}
		m.mu.Unlock()
	}
}
