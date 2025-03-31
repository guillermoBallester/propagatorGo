package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"propagatorGo/internal/config"
	"propagatorGo/internal/constants"
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

// EnqueueTask adds a task to the appropriate queue
func (s *Service) EnqueueTask(ctx context.Context, task *Task) error {
	queueName := QueueName(task.Type)
	return s.queueSvc.Enqueue(ctx, queueName, task)
}

// EnqueueStocks adds all enabled stock symbols as tasks for a specific task type
func (s *Service) EnqueueStocks(ctx context.Context, taskType string, source string) error {
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

		task := NewTask(taskType)
		task.SetParam("symbol", stock.Symbol)
		task.SetParam("source", source)

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

// CreateScrapeTask creates a new scrape task for a given symbol and source
func (s *Service) CreateScrapeTask(symbol, source string) *Task {
	task := NewTask(constants.TaskTypeScrape)
	task.SetParam("symbol", symbol)
	task.SetParam("source", source)
	return task
}

// CreateConsumeTask creates a new consume task with article data
func (s *Service) CreateConsumeTask(symbol, source string, article interface{}) *Task {
	task := NewTask(constants.TaskTypeConsume)
	task.SetParam("symbol", symbol)
	task.SetParam("source", source)
	task.SetParam("article", article)
	return task
}
