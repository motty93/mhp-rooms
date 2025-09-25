package sse

import (
	"context"
	"fmt"
	"time"

	redisinfra "mhp-rooms/internal/infrastructure/redis"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type RedisTokenManager struct {
	redis redisinfra.RedisClient
	ttl   time.Duration
}

func NewRedisTokenManager(redisClient redisinfra.RedisClient, ttl time.Duration) *RedisTokenManager {
	return &RedisTokenManager{
		redis: redisClient,
		ttl:   ttl,
	}
}

func (m *RedisTokenManager) GenerateToken(roomID, userID uuid.UUID) (string, error) {
	token := generateSecureToken()

	info := TokenInfo{
		RoomID:    roomID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	key := fmt.Sprintf("sse:token:%s", token)
	ctx := context.Background()

	if err := m.redis.Set(ctx, key, info, m.ttl); err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	return token, nil
}

func (m *RedisTokenManager) ConsumeToken(token string) (*TokenInfo, error) {
	key := fmt.Sprintf("sse:token:%s", token)
	ctx := context.Background()

	var info TokenInfo
	if err := m.redis.GetStruct(ctx, key, &info); err != nil {
		if err == goredis.Nil || err == redisinfra.Nil {
			return nil, ErrTokenNotFound
		}

		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// One-time use: トークン消費後に削除（失敗しても処理は続行）
	if err := m.redis.Delete(ctx, key); err != nil {
		fmt.Printf("Failed to delete consumed token: %v\n", err)
	}

	return &info, nil
}
