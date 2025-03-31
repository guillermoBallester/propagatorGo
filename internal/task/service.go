package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"propagatorGo/internal/config"
)

// QueueService defines the minimum interface needed for task queue operations
type QueueService interface {
	Enqueue(ctx context.Context, queueName string, task interface{}) error
	Dequeue(ctx context.Context, queueName string, timeout int) ([]byte, error)
	QueueLength(ctx context.Context, queueName string) (int64, error)
	ClearQueue(ctx context.Context, queueName string) error
}

// Service manages task processing
type Service struct {
	config   *config.Config
	queueSvc QueueService
}

// NewService creates a new task service
func NewService(cfg *config.Config, queue QueueService) *Service {
	return &Service{
		config:   cfg,
		queueSvc: queue,
	}
}

// QueueName constructs the appropriate queue name for a task type
func QueueName(taskType string) string {
	return fmt.Sprintf("task:%s", taskType)
}

// EnqueueAll adds all enabled items for a specific task type
func (s *Service) EnqueueAll(ctx context.Context, taskType string, source string) error {
	queueName := QueueName(taskType)

	// Check if queue already has items
	length, err := s.queueSvc.QueueLength(ctx, queueName)
	if err != nil {
		return fmt.Errorf("failed to check queue length: %w", err)
	}

	if length > 0 {
		log.Printf("Queue %s already has %d items, skipping initialization", queueName, length)
		return nil
	}

	// Get enabled stocks from config (this still refers to stocks but the logic is generic)
	var tasksAdded int
	for _, stock := range s.config.StockList.Stocks {
		if !stock.Enabled {
			continue
		}
		var task *Task
		task = NewTask(stock.Symbol, taskType, source)
		if err := s.queueSvc.Enqueue(ctx, queueName, task); err != nil {
			log.Printf("Error enqueueing %s task for %s: %v", taskType, stock.Symbol, err)
			continue
		}

		tasksAdded++
	}

	log.Printf("Added %d tasks to %s queue", tasksAdded, queueName)
	return nil
}

// GetNext retrieves the next task from the queue
func (s *Service) GetNext(ctx context.Context, taskType string, timeout int) (*Task, error) {
	queueName := QueueName(taskType)
	fmt.Printf("Getting next task for %s queue", queueName)
	data, err := s.queueSvc.Dequeue(ctx, queueName, timeout)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Got task %s", data)

	if data == nil {
		return nil, nil
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}
