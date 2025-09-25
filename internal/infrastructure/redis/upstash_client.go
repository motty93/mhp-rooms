package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type UpstashClient struct {
	client        *goredis.Client
	maxRetries    int
	retryInterval time.Duration
}

// コンパイル時のインターフェース実装確認
var _ RedisClient = (*UpstashClient)(nil)

func NewUpstashClient(url string, token string) (RedisClient, error) {
	var client *goredis.Client

	if token != "" {
		// Upstash用の認証設定
		opts, err := goredis.ParseURL(url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}

		opts.Password = token
		client = goredis.NewClient(opts)
	} else {
		// 通常のRedis接続
		opts, err := goredis.ParseURL(url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}

		client = goredis.NewClient(opts)
	}

	// Test connection with retry
	ctx := context.Background()
	if err := retryWithBackoff(ctx, 3, func() error {
		return client.Ping(ctx).Err()
	}); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis/Upstash: %w", err)
	}

	return &UpstashClient{
		client:        client,
		maxRetries:    3,
		retryInterval: 1 * time.Second,
	}, nil
}

func (c *UpstashClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return retryWithBackoff(ctx, c.maxRetries, func() error {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}

		return c.client.Set(ctx, key, data, ttl).Err()
	})
}

func (c *UpstashClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *UpstashClient) GetStruct(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (c *UpstashClient) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *UpstashClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return retryWithBackoff(ctx, c.maxRetries, func() error {
		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		return c.client.Publish(ctx, channel, data).Err()
	})
}

func (c *UpstashClient) Subscribe(ctx context.Context, channels ...string) *goredis.PubSub {
	return c.client.Subscribe(ctx, channels...)
}

func (c *UpstashClient) Close() error {
	return c.client.Close()
}

func (c *UpstashClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// SubscribeWithReconnect creates a subscription with automatic reconnection
// これはUpstashClientの追加機能（インターフェースには含まれない）
func (c *UpstashClient) SubscribeWithReconnect(ctx context.Context, channels ...string) <-chan *goredis.Message {
	output := make(chan *goredis.Message, 100)

	go func() {
		defer close(output)
		retryInterval := c.retryInterval

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Create subscription
			pubsub := c.client.Subscribe(ctx, channels...)
			ch := pubsub.Channel()

			log.Printf("Subscribed to channels: %v", channels)

			// Process messages
		innerLoop:
			for {
				select {
				case <-ctx.Done():
					pubsub.Close()
					return

				case msg, ok := <-ch:
					if !ok {
						// Channel closed, need to reconnect
						log.Printf("Subscription channel closed, reconnecting...")
						pubsub.Close()

						// Exponential backoff before reconnect
						time.Sleep(retryInterval)
						retryInterval = time.Duration(math.Min(float64(retryInterval*2), float64(30*time.Second)))
						break innerLoop
					}

					if msg != nil {
						// Reset retry interval on successful message
						retryInterval = c.retryInterval

						select {
						case output <- msg:
						case <-ctx.Done():
							pubsub.Close()
							return
						default:
							// Output channel full, drop message
							log.Printf("Warning: dropping message due to full channel")
						}
					}
				}
			}
		}
	}()

	return output
}

// ConsumeToken implements atomic token consumption for SSE
// これはUpstashClientの追加機能（インターフェースには含まれない）
func (c *UpstashClient) ConsumeToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("sse:token:%s", token)

	// GETDEL is atomic get-and-delete (Redis 6.2+)
	// Upstash supports this command
	result, err := c.client.GetDel(ctx, key).Result()
	if err == goredis.Nil {
		return "", ErrTokenNotFound
	}
	if err != nil {
		// Fallback to transaction for older versions or compatibility issues
		pipe := c.client.TxPipeline()
		getCmd := pipe.Get(ctx, key)
		pipe.Del(ctx, key)

		_, err = pipe.Exec(ctx)
		if err != nil {
			return "", err
		}

		result, err = getCmd.Result()
		if err == goredis.Nil {
			return "", ErrTokenNotFound
		}
		if err != nil {
			return "", err
		}
	}

	return result, nil
}

// SetTokenNX sets a token with NX (only if not exists) and TTL
// これはUpstashClientの追加機能（インターフェースには含まれない）
func (c *UpstashClient) SetTokenNX(ctx context.Context, token string, value interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("sse:token:%s", token)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal token value: %w", err)
	}

	// SET with NX and EX options
	// NX: Only set the key if it does not already exist
	// EX: Set the specified expire time, in seconds
	ok, err := c.client.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to set token: %w", err)
	}

	if !ok {
		return fmt.Errorf("token already exists")
	}

	return nil
}

func (c *UpstashClient) SetWithRetry(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl)
}

func (c *UpstashClient) PublishWithRetry(ctx context.Context, channel string, message interface{}) error {
	return c.Publish(ctx, channel, message)
}

func retryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	backoff := 100 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err

			// Log retry attempts for debugging
			if i < maxRetries-1 {
				log.Printf("Retry attempt %d/%d failed: %v. Retrying in %v...",
					i+1, maxRetries, err, backoff)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					// Exponential backoff with cap at 10 seconds
					backoff = time.Duration(math.Min(float64(backoff*2), float64(10*time.Second)))
				}
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

var (
	ErrTokenNotFound = goredis.Nil
)
