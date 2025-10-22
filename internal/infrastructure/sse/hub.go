package sse

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Event はSSEで送信するイベント
type Event struct {
	ID   string      `json:"id"`
	Type string      `json:"type"` // message, member_join, member_leave, room_update
	Data interface{} `json:"data"`
}

// Client はSSE接続を表す
type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID
	RoomID uuid.UUID
	Send   chan Event
}

// Hub は部屋ごとのSSE接続を管理
type Hub struct {
	rooms      map[uuid.UUID]map[uuid.UUID]*Client // roomID -> userID -> client
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	mu         sync.RWMutex
}

// BroadcastMessage はブロードキャストするメッセージ
type BroadcastMessage struct {
	RoomID uuid.UUID
	Event  Event
}

// NewHub は新しいHubを作成
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage),
	}
}

// Run はHubのメインループ
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.RoomID]; !ok {
				h.rooms[client.RoomID] = make(map[uuid.UUID]*Client)
			}
			h.rooms[client.RoomID][client.UserID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.RoomID]; ok {
				if _, ok := room[client.UserID]; ok {
					delete(room, client.UserID)
					close(client.Send)
					if len(room) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if room, ok := h.rooms[message.RoomID]; ok {
				for _, client := range room {
					select {
					case client.Send <- message.Event:
					default:
						// クライアントのバッファがいっぱいの場合はスキップ
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToRoom は特定の部屋にイベントをブロードキャスト
func (h *Hub) BroadcastToRoom(roomID uuid.UUID, event Event) {
	h.broadcast <- BroadcastMessage{
		RoomID: roomID,
		Event:  event,
	}
}

// Register はクライアントを登録
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister はクライアントを登録解除
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// SerializeEvent はイベントをSSE形式にシリアライズ
func SerializeEvent(event Event) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("id: %s\nevent: %s\ndata: %s\n\n",
		event.ID, event.Type, string(data)), nil
}
