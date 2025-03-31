// scraper_worker.go
package worker

import (
	"context"
	"fmt"
	"log"
	"propagatorGo/internal/constants"
	scraper "propagatorGo/internal/scrapper"
	"propagatorGo/internal/task"
	"time"
)

// ScraperPublisherWorker scrapes websites and publishes to Redis
type ScraperPublisherWorker struct {
	BaseWorker
	scraperService *scraper.Service
	taskService    *task.Service
}

// NewScraperWorker creates a new scraper worker
func NewScraperWorker(bw BaseWorker, scraperSvc *scraper.Service, taskSvc *task.Service) *ScraperPublisherWorker {
	return &ScraperPublisherWorker{
		BaseWorker:     bw,
		scraperService: scraperSvc,
		taskService:    taskSvc,
	}
}

// Start begins the scraping process
func (w *ScraperPublisherWorker) Start(ctx context.Context) error {
	if !w.SetActive(true) {
		return fmt.Errorf("worker %s is already running", w.Name())
	}

	for w.IsActive() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			w.Stats.RecordStart()

			fmt.Println("Trying to get next task")
			nextTask, err := w.taskService.GetNext(ctx, constants.WorkerTypeScraper, 5)
			if err != nil {
				log.Printf("Error getting task: %v", err)
				w.Stats.RecordItemFailed()
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Printf("Got next task %s", nextTask)

			// If no task returned within timeout, try again
			if nextTask == nil {
				continue
			}

			log.Printf("Worker %s processing symbol: %s from source %s",
				w.Name(), nextTask.Params.Symbol, nextTask.Params.Source)

			articles, err := w.scraperService.ScrapeAndPublish(
				ctx,
				nextTask.Params.Source,
				nextTask.Params.Symbol,
			)
			if err != nil {
				log.Printf("Error processing symbol %s: %v", nextTask.Params.Symbol, err)
				w.Stats.RecordItemFailed()
				continue
			}

			w.Stats.RecordItemProcessed()
			stats := w.Stats.GetSnapshot()
			log.Printf("[%s] Task completed for %s. Articles: %d, Total processed: %d, Successful: %d, Failed: %d",
				w.Name(),
				nextTask.Params.Symbol,
				len(articles),
				stats.ItemsProcessed,
				stats.ItemsSuccessful,
				stats.ItemsFailed)
		}
	}

	return nil
}
