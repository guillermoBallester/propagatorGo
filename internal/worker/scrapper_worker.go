// scraper_worker.go
package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	scraper "github.com/guillermoballester/propagatorGo/internal/scrapper"
)

// ScraperWorker scrapes websites and publishes to Redis
type ScraperWorker struct {
	BaseWorker
	scraperService *scraper.Service
	source         string
	WorkManager    *WorkManager
}

// NewScraperWorker creates a new scraper worker
func NewScraperWorker(bw BaseWorker, scraperSvc *scraper.Service, wm *WorkManager, source string) *ScraperWorker {
	return &ScraperWorker{
		BaseWorker:     bw,
		scraperService: scraperSvc,
		WorkManager:    wm,
		source:         source,
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

			stock := w.WorkManager.GetNextStock()
			if stock == nil {
				log.Printf("No stocks available for worker %s", w.Name())
				time.Sleep(5 * time.Second)
				continue
			}

			symbol := stock.Symbol
			source := w.source

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
