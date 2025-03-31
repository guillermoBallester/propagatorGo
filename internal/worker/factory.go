package worker

import (
	"fmt"
	"propagatorGo/internal/constants"
	"propagatorGo/internal/repository"
	scraper "propagatorGo/internal/scrapper"
	"propagatorGo/internal/task"
)

// Factory creates workers based on configuration
type Factory struct {
	scraperService *scraper.Service
	taskService    *task.Service
	repository     *repository.ArticleRepository
}

// NewWorkerFactory creates a new worker factory
func NewWorkerFactory(scraperSvc *scraper.Service, taskSvc *task.Service, repo *repository.ArticleRepository) *Factory {
	return &Factory{
		scraperService: scraperSvc,
		taskService:    taskSvc,
		repository:     repo,
	}
}

// CreateWorker creates a worker of the specified type
func (f *Factory) CreateWorker(id int, workerType string) (Worker, error) {
	baseName := fmt.Sprintf("%s%d", workerType, id)
	baseWorker := NewBaseWorker(id, baseName, workerType)

	switch workerType {
	case constants.WorkerTypeScraper:
		return NewScraperWorker(baseWorker, f.scraperService, f.taskService), nil
	case constants.WorkerTypeConsumer:
		return NewConsumerWorker(baseWorker, f.taskService, f.repository), nil
	default:
		return nil, fmt.Errorf("unknown worker type: %s", workerType)
	}
}
