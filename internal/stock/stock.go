package stock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"propagatorGo/internal/config"
)

// TaskType defines which kind of stock task it is
type TaskType string

const (
	TaskTypeScrape  TaskType = "scrape"
	TaskTypeAPICall TaskType = "api_call"
	// TODO More tasks?
)

// Task represents a generic stock-related task
type Task struct {
	Type     TaskType          `json:"type"`
	Stock    config.Stock      `json:"stock"`
	Source   string            `json:"source,omitempty"` // e.g., "yahoo", "api:alphavantage"
	Params   map[string]string `json:"params,omitempty"`
	Priority int               `json:"priority,omitempty"` // Lower number = higher priority
}

// Service is the main service for stock task management
type Service struct {
	config   *config.Config
	queueSvc QueueService
}

// QueueService defines the interface for queue operations
type QueueService interface {
	Enqueue(ctx context.Context, queueName string, task interface{}) error
	Dequeue(ctx context.Context, queueName string, timeout int) ([]byte, error)
	QueueLength(ctx context.Context, queueName string) (int64, error)
	ClearQueue(ctx context.Context, queueName string) error
}

// NewService creates a new stock service
func NewService(cfg *config.Config, queue QueueService) *Service {
	return &Service{
		config:   cfg,
		queueSvc: queue,
	}
}

// QueueName constructs the appropriate queue name for a task type
func QueueName(taskType TaskType) string {
	return fmt.Sprintf("stock_tasks:%s", taskType)
}

// EnqueueAllStocks adds all enabled stocks for a specific task type
func (s *Service) EnqueueAllStocks(ctx context.Context, taskType TaskType, source string) error {
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

	// Get enabled stocks from config
	var tasksAdded int
	for _, stock := range s.config.StockList.Stocks {
		if !stock.Enabled {
			continue
		}

		task := Task{
			Type:   taskType,
			Stock:  stock,
			Source: source,
		}

		if err := s.queueSvc.Enqueue(ctx, queueName, task); err != nil {
			log.Printf("Error enqueueing %s task for %s: %v", taskType, stock.Symbol, err)
			continue
		}

		tasksAdded++
	}

	log.Printf("Added %d tasks to %s queue", tasksAdded, queueName)
	return nil
}

// GetNextTask retrieves the next task from the queue
func (s *Service) GetNextTask(ctx context.Context, taskType TaskType, timeout int) (*Task, error) {
	queueName := QueueName(taskType)

	data, err := s.queueSvc.Dequeue(ctx, queueName, timeout)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// FilterStocks returns stocks that match certain criteria
func (s *Service) FilterStocks(onlyEnabled bool) []config.Stock {
	var result []config.Stock

	for _, stock := range s.config.StockList.Stocks {
		if onlyEnabled && !stock.Enabled {
			continue
		}

		result = append(result, stock)
	}

	return result
}
