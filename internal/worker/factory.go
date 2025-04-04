package worker

import (
	"fmt"

	"github.com/guillermoballester/propagatorGo/internal/config"
	"github.com/guillermoballester/propagatorGo/internal/constants"
	"github.com/guillermoballester/propagatorGo/internal/repository"
	scraper "github.com/guillermoballester/propagatorGo/internal/scrapper"
	"github.com/guillermoballester/propagatorGo/internal/task"
)

// Factory creates workers based on configuration
type Factory struct {
	scraperService *scraper.Service
	taskService    *task.Service
	repository     *repository.ArticleRepository
	workManager    *WorkManager
}

// NewWorkerFactory creates a new worker factory
func NewWorkerFactory(cfg *config.Config, scraperSvc *scraper.Service, taskSvc *task.Service, repo *repository.ArticleRepository) *Factory {
	return &Factory{
		scraperService: scraperSvc,
		taskService:    taskSvc,
		repository:     repo,
		workManager:    NewWorkManagerFromConfig(cfg),
	}
}

// CreateWorker creates a worker of the specified type
func (f *Factory) CreateWorker(id int, workerType string, source string) (Worker, error) {
	baseName := fmt.Sprintf("%s%d", workerType, id)
	baseWorker := NewBaseWorker(id, baseName, workerType)

	switch workerType {
	case constants.WorkerTypeScraper:
		return NewScraperWorker(baseWorker, f.scraperService, f.workManager, source), nil
	case constants.WorkerTypeConsumer:
		return NewConsumerWorker(baseWorker, f.taskService, f.repository), nil
	default:
		return nil, fmt.Errorf("unknown worker type: %s", workerType)
	}
}
