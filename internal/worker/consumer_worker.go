// consumer_worker.go
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"propagatorGo/internal/database"
	"propagatorGo/internal/queue"
	scraper "propagatorGo/internal/scrapper"
	"sync/atomic"
	"time"
)

// ConsumerWriterWorker consumes messages from Redis and stores in the database
type ConsumerWriterWorker struct {
	BaseWorker
	redis        *queue.RedisClient
	queueName    string
	pollInterval time.Duration
}

// NewConsumerWorker creates a new consumer worker
func NewConsumerWorker(id int, redis *queue.RedisClient, db *database.PostgresClient, queueName string) *ConsumerWriterWorker {
	return &ConsumerWriterWorker{
		redis:        redis,
		queueName:    queueName,
		pollInterval: 1 * time.Second,
	}
}

// Start begins consuming messages from Redis
func (w *ConsumerWriterWorker) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&w.Active, 0, 1) {
		return fmt.Errorf("worker %s is already running", w.Name)
	}

	log.Printf("Consumer worker %s started, polling queue %s", w.Name, w.queueName)
	for atomic.LoadInt32(&w.Active) == 1 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Try to consume a message
			message, err := w.redis.ConsumeMessage(ctx, w.queueName, w.pollInterval)
			if err != nil {
				log.Printf("Error consuming message: %v", err)
				time.Sleep(1 * time.Second) // Backoff on error
				continue
			}

			if message == nil {
				// No message available, try again
				continue
			}

			// Process the message
			/*if err := w.processMessage(ctx, message); err != nil {
				log.Printf("Error processing message: %v", err)
			}*/
		}
	}

	return nil
}

// processMessage handles a single message from Redis
func (w *ConsumerWriterWorker) processMessage(ctx context.Context, data []byte) error {
	var article scraper.ArticleData
	if err := json.Unmarshal(data, &article); err != nil {
		return fmt.Errorf("error unmarshaling article: %w", err)
	}

	// TODO: Store the article in the database

	log.Printf("Processed article: %s", article.Title)
	return nil
}

// Stop gracefully stops the worker
func (w *ConsumerWriterWorker) Stop() error {
	atomic.StoreInt32(&w.Active, 0)
	return nil
}

// Name returns the worker name
func (w *ConsumerWriterWorker) Name() string {
	return w.BaseWorker.Name
}
