package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	redisinfra "mhp-rooms/internal/infrastructure/redis"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

var (
	globalEventBus EventBus
	eventBusMu     sync.RWMutex
)

type MessageEvent struct {
	RoomID    uuid.UUID       `json:"room_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

type EventBus interface {
	PublishMessage(ctx context.Context, roomID uuid.UUID, event Event) error
	Subscribe(ctx context.Context, hub *Hub) error
	Close() error
}

type LocalEventBus struct {
	hub *Hub
}

func NewLocalEventBus(hub *Hub) *LocalEventBus {
	return &LocalEventBus{
		hub: hub,
	}
}

func (b *LocalEventBus) PublishMessage(ctx context.Context, roomID uuid.UUID, event Event) error {
	b.hub.BroadcastToRoom(roomID, event)
	return nil
}

func (b *LocalEventBus) Subscribe(ctx context.Context, hub *Hub) error {
	// No-op for local event bus
	return nil
}

func (b *LocalEventBus) Close() error {
	// No-op for local event bus
	return nil
}

type RedisEventBus struct {
	redis      redisinfra.RedisClient
	hub        *Hub
	stopChan   chan struct{}
	wg         sync.WaitGroup
	subscribed bool
	mu         sync.Mutex
}

func NewRedisEventBus(redisClient redisinfra.RedisClient, hub *Hub) *RedisEventBus {
	return &RedisEventBus{
		redis:    redisClient,
		hub:      hub,
		stopChan: make(chan struct{}),
	}
}

func (b *RedisEventBus) PublishMessage(ctx context.Context, roomID uuid.UUID, event Event) error {
	payload, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	messageEvent := MessageEvent{
		RoomID:    roomID,
		Type:      event.Type,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	channel := fmt.Sprintf("room:%s:messages", roomID.String())

	if err := b.redis.Publish(ctx, channel, messageEvent); err != nil {
		return fmt.Errorf("failed to publish message event: %w", err)
	}

	return nil
}

func (b *RedisEventBus) Subscribe(ctx context.Context, hub *Hub) error {
	b.mu.Lock()
	if b.subscribed {
		b.mu.Unlock()
		return fmt.Errorf("already subscribed")
	}

	b.subscribed = true
	b.hub = hub
	b.mu.Unlock()

	// ワイルドカード購読: すべてのルームメッセージを購読
	pubsub := b.redis.Subscribe(ctx, "room:*:messages")

	b.wg.Add(1)
	go b.processMessages(ctx, pubsub)

	log.Println("Redis event bus subscription started")

	return nil
}

func (b *RedisEventBus) Close() error {
	close(b.stopChan)
	b.wg.Wait()
	return nil
}

func InitializeEventBus(redisURL string, hub *Hub) error {
	eventBusMu.Lock()
	defer eventBusMu.Unlock()

	if redisURL != "" {
		redisClient, err := redisinfra.NewClient(redisURL)
		if err != nil {
			return fmt.Errorf("failed to create Redis client for event bus: %w", err)
		}

		eventBus := NewRedisEventBus(redisClient, hub)

		ctx := context.Background()
		if err := eventBus.Subscribe(ctx, hub); err != nil {
			return fmt.Errorf("failed to subscribe to Redis events: %w", err)
		}

		globalEventBus = eventBus
		fmt.Println("Using Redis event bus")
	} else {
		globalEventBus = NewLocalEventBus(hub)
		fmt.Println("Using local event bus")
	}

	return nil
}

// InitializeEventBusWithClient DI用: 既存のRedisクライアントを使用
func InitializeEventBusWithClient(redisClient redisinfra.RedisClient, hub *Hub) error {
	eventBusMu.Lock()
	defer eventBusMu.Unlock()

	if redisClient != nil {
		eventBus := NewRedisEventBus(redisClient, hub)

		ctx := context.Background()
		if err := eventBus.Subscribe(ctx, hub); err != nil {
			return fmt.Errorf("failed to subscribe to Redis events: %w", err)
		}

		globalEventBus = eventBus
		fmt.Println("Using Redis event bus with provided client")
	} else {
		globalEventBus = NewLocalEventBus(hub)
		fmt.Println("Using local event bus")
	}

	return nil
}

func GetEventBus() EventBus {
	eventBusMu.RLock()
	defer eventBusMu.RUnlock()

	return globalEventBus
}

func SetEventBus(bus EventBus) {
	eventBusMu.Lock()
	defer eventBusMu.Unlock()

	globalEventBus = bus
}

func (b *RedisEventBus) processMessages(ctx context.Context, pubsub *goredis.PubSub) {
	defer b.wg.Done()
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-b.stopChan:
			log.Println("Stopping Redis event bus subscription")
			return

		case <-ctx.Done():
			log.Println("Context cancelled, stopping Redis event bus")
			return

		case msg := <-ch:
			if msg == nil {
				continue
			}

			var messageEvent MessageEvent
			if err := json.Unmarshal([]byte(msg.Payload), &messageEvent); err != nil {
				log.Printf("Failed to unmarshal message event: %v", err)
				continue
			}

			var data interface{}
			if err := json.Unmarshal(messageEvent.Payload, &data); err != nil {
				log.Printf("Failed to unmarshal event payload: %v", err)
				continue
			}

			event := Event{
				ID:   uuid.New().String(),
				Type: messageEvent.Type,
				Data: data,
			}

			b.hub.BroadcastToRoom(messageEvent.RoomID, event)
		}
	}
}
