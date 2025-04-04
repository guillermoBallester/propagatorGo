package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/guillermoballester/propagatorGo/internal/config"

	"github.com/go-redis/redis/v8"
)

// RedisClient handles communication with Redis
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Address,
		Password:        cfg.Password,
		DB:              0,
		MaxRetries:      5,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 2 * time.Second,
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

// Enqueue adds an item to a Redis queue
func (r *RedisClient) Enqueue(ctx context.Context, queueName string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = r.client.RPush(ctx, queueName, data).Err()
	if err != nil {
		return fmt.Errorf("failed to push task to queue '%s': %w", queueName, err)
	}

	return nil
}

// Dequeue retrieves an item from a Redis queue with a timeout
func (r *RedisClient) Dequeue(ctx context.Context, queueName string, timeoutSeconds int) ([]byte, error) {
	timeout := time.Duration(timeoutSeconds) * time.Second

	result, err := r.client.BLPop(ctx, timeout, queueName).Result()
	if err != nil {
		// If timeout or nil, return nil without error
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	// BLPOP returns [queueName, value]
	if len(result) < 2 {
		return nil, nil
	}

	return []byte(result[1]), nil
}

// QueueLength returns the number of items in a queue
func (r *RedisClient) QueueLength(ctx context.Context, queueName string) (int64, error) {
	return r.client.LLen(ctx, queueName).Result()
}

// ClearQueue removes all items from a queue
func (r *RedisClient) ClearQueue(ctx context.Context, queueName string) error {
	return r.client.Del(ctx, queueName).Err()
}

// IsQueueEmpty checks if a queue is empty
func (r *RedisClient) IsQueueEmpty(ctx context.Context, queueName string) (bool, error) {
	length, err := r.client.LLen(ctx, queueName).Result()
	if err != nil {
		return false, err
	}

	return length == 0, nil
}
