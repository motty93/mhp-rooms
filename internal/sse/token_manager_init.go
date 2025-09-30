package sse

import (
	"fmt"
	"time"

	redisinfra "mhp-rooms/internal/infrastructure/redis"
)

func InitializeTokenManager(redisURL string, ttl time.Duration) error {
	if redisURL != "" {
		redisClient, err := redisinfra.NewClient(redisURL)
		if err != nil {
			return fmt.Errorf("failed to create Redis client: %w", err)
		}

		SetTokenManager(NewRedisTokenManager(redisClient, ttl))
		fmt.Println("Using Redis token manager")
	} else {
		SetTokenManager(NewInMemoryTokenManager(ttl))
		fmt.Println("Using in-memory token manager")
	}

	return nil
}

// InitializeTokenManagerWithClient DI用: 既存のRedisクライアントを使用する場合
func InitializeTokenManagerWithClient(redisClient redisinfra.RedisClient, ttl time.Duration) error {
	if redisClient != nil {
		SetTokenManager(NewRedisTokenManager(redisClient, ttl))
		fmt.Println("Using Redis token manager with provided client")
	} else {
		SetTokenManager(NewInMemoryTokenManager(ttl))
		fmt.Println("Using in-memory token manager")
	}

	return nil
}
