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

// ScraperWorker scrapes websites and publishes to Redis
type ScraperWorker struct {
	BaseWorker
	scraperService *scraper.Service
	taskService    *task.Service
}

// NewScraperWorker creates a new scraper worker
func NewScraperWorker(bw BaseWorker, scraperSvc *scraper.Service, taskSvc *task.Service) *ScraperWorker {
	return &ScraperWorker{
		BaseWorker:     bw,
		scraperService: scraperSvc,
		taskService:    taskSvc,
	}
}

// Start begins the scraping process
func (w *ScraperWorker) Start(ctx context.Context) error {
	if !w.SetActive(true) {
		return fmt.Errorf("worker %s is already running", w.Name())
	}

	for w.IsActive() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			w.Stats.RecordStart()

			nextTask, err := w.taskService.GetNext(ctx, constants.TaskTypeScrape, 5)
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

			symbol, err := nextTask.GetParamString("symbol")
			if err != nil {
				log.Printf("Error getting symbol from task: %v", err)
				w.Stats.RecordItemFailed()
				continue
			}

			source, err := nextTask.GetParamString("source")
			if err != nil {
				log.Printf("Error getting source from task: %v", err)
				w.Stats.RecordItemFailed()
				continue
			}

			log.Printf("Worker %s processing symbol: %s from source %s",
				w.Name(), symbol, source)

			articles, err := w.scraperService.ScrapeAndPublish(
				ctx,
				source,
				symbol,
			)
			if err != nil {
				log.Printf("Error processing symbol %s: %v", symbol, err)
				w.Stats.RecordItemFailed()
				continue
			}

			w.Stats.RecordItemProcessed()
			stats := w.Stats.GetSnapshot()
			log.Printf("[%s] Task completed for %s. Articles: %d, Total processed: %d, Successful: %d, Failed: %d",
				w.Name(),
				symbol,
				len(articles),
				stats.ItemsProcessed,
				stats.ItemsSuccessful,
				stats.ItemsFailed)
		}
	}

	return nil
}
