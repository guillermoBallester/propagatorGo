package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"propagatorGo/internal/config"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient handles communication with Redis
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: client,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// PublishMessage publishes a message to a Redis queue
func (r *RedisClient) PublishMessage(ctx context.Context, queueName string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.client.RPush(ctx, queueName, data).Err()
	if err != nil {
		return fmt.Errorf("failed to push message to queue '%s': %w", queueName, err)
	}

	return nil
}

// ConsumeMessage retrieves a message from a Redis queue
func (r *RedisClient) ConsumeMessage(ctx context.Context, queueName string, timeout time.Duration) ([]byte, error) {
	result, err := r.client.BLPop(ctx, timeout, queueName).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // No message available
		}
		return nil, err
	}

	// BLPOP returns [queueName, value]
	if len(result) < 2 {
		return nil, nil
	}

	return []byte(result[1]), nil
}
