package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// SSEToken は一時的なSSE接続用トークンを表す
type SSEToken struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	RoomID    uuid.UUID `json:"room_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SSETokenManager はSSE用の一時トークンを管理する
type SSETokenManager struct {
	tokens map[string]*SSEToken
	mutex  sync.RWMutex
}

func NewSSETokenManager() *SSETokenManager {
	manager := &SSETokenManager{
		tokens: make(map[string]*SSEToken),
	}

	// 期限切れトークンを定期的にクリーンアップ
	go manager.cleanup()

	return manager
}

func (m *SSETokenManager) GenerateToken(userID, roomID uuid.UUID) string {
	// 32バイトのランダムトークンを生成
	bytes := make([]byte, 32)
	rand.Read(bytes)
	token := hex.EncodeToString(bytes)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.tokens[token] = &SSEToken{
		Token:     token,
		UserID:    userID,
		RoomID:    roomID,
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5分間有効
	}

	return token
}

func (m *SSETokenManager) ValidateToken(token string) (*SSEToken, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sseToken, exists := m.tokens[token]
	if !exists || time.Now().After(sseToken.ExpiresAt) {
		return nil, false
	}

	return sseToken, true
}

func (m *SSETokenManager) ConsumeToken(token string) (*SSEToken, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sseToken, exists := m.tokens[token]
	if !exists || time.Now().After(sseToken.ExpiresAt) {
		return nil, false
	}

	// トークンを一度だけ使用可能にする（使用後削除）
	delete(m.tokens, token)
	return sseToken, true
}

func (m *SSETokenManager) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mutex.Lock()
			now := time.Now()
			for token, sseToken := range m.tokens {
				if now.After(sseToken.ExpiresAt) {
					delete(m.tokens, token)
				}
			}
			m.mutex.Unlock()
		}
	}
}

// グローバルなSSEトークンマネージャー
var globalSSETokenManager = NewSSETokenManager()

type SSETokenHandler struct {
	BaseHandler
}

func NewSSETokenHandler(repo *repository.Repository) *SSETokenHandler {
	return &SSETokenHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
	}
}

// GenerateSSEToken はSSE接続用の一時トークンを生成する
func (h *SSETokenHandler) GenerateSSEToken(w http.ResponseWriter, r *http.Request) {
	// URLパラメータから部屋IDを取得
	roomIDStr := chi.URLParam(r, "id")

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "無効な部屋IDです", http.StatusBadRequest)
		return
	}

	// 認証チェック（DBユーザー情報を取得）
	user, ok := middleware.GetDBUserFromContext(r.Context())
	if !ok || user == nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// 部屋のメンバーチェック
	if !h.repo.Room.IsUserJoinedRoom(roomID, user.ID) {
		http.Error(w, "部屋のメンバーではありません", http.StatusForbidden)
		return
	}

	// 一時トークンを生成
	token := globalSSETokenManager.GenerateToken(user.ID, roomID)

	response := map[string]string{
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
