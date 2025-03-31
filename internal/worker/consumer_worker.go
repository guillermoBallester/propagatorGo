// consumer_worker.go
package worker

import (
	"context"
	"fmt"
	"log"
	"propagatorGo/internal/constants"
	"propagatorGo/internal/database"
	"propagatorGo/internal/repository"
	"propagatorGo/internal/task"
	"time"
)

// ConsumerWorker consumes messages from Redis and stores in the database
type ConsumerWorker struct {
	BaseWorker
	taskService *task.Service
	repository  *repository.ArticleRepository
}

// NewConsumerWorker creates a new consumer worker
func NewConsumerWorker(bw BaseWorker, taskSvc *task.Service) *ConsumerWorker {
	return &ConsumerWorker{
		BaseWorker:  bw,
		taskService: taskSvc,
	}
}

// Start begins consuming messages from Redis
func (w *ConsumerWorker) Start(ctx context.Context) error {
	if !w.SetActive(true) {
		return fmt.Errorf("worker %s is already running", w.Name())
	}

	for w.IsActive() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			w.Stats.RecordStart()

			nextTask, err := w.taskService.GetNext(ctx, constants.TaskTypeConsume, 5)
			if err != nil {
				log.Printf("Error getting task: %v", err)
				w.Stats.RecordItemFailed()
				time.Sleep(1 * time.Second)
				continue
			}

			// If no task returned within timeout, try again
			if nextTask == nil {
				continue
			}

			log.Printf("Worker %s processing article from source %s for symbol %s",
				w.Name(), nextTask.Params.Source, nextTask.Params.Symbol)

			// Extract article from task
			article, err := nextTask.GetArticle()
			if err != nil {
				log.Printf("Error extracting article: %v", err)
				w.Stats.RecordItemFailed()
				continue
			}

			// Convert to database model
			dbArticle := database.Article{
				Title:     article.Title,
				URL:       article.URL,
				Text:      article.Text,
				SiteName:  article.SiteName,
				ScrapedAt: article.ScrapedAt,
				Symbol:    nextTask.Params.Symbol,
			}

			err = w.repository.SaveArticle(ctx, dbArticle)
			if err != nil {
				log.Printf("Error saving article to database: %v", err)
				w.Stats.RecordItemFailed()
				continue
			}

			w.Stats.RecordItemProcessed()
			stats := w.Stats.GetSnapshot()
			log.Printf("[%s] Task completed for %s. Articles: %d, Total processed: %d, Successful: %d, Failed: %d",
				w.Name(),
				nextTask.Params.Symbol,
				stats.ItemsProcessed,
				stats.ItemsSuccessful,
				stats.ItemsFailed)
		}

	}
	return nil
}
