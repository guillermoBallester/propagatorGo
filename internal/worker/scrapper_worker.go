// scraper_worker.go
package worker

import (
	"context"
	"fmt"
	"log"
	scraper "propagatorGo/internal/scrapper"
	"sync/atomic"
	"time"
)

// ScraperPublisherWorker scrapes websites and publishes to Redis
type ScraperPublisherWorker struct {
	BaseWorker
	scraper   *scraper.NewsScraper
	publisher *scraper.ArticlePublisher
}

// NewScraperWorker creates a new scraper worker
func NewScraperWorker(bw BaseWorker, scraper *scraper.NewsScraper, publisher *scraper.ArticlePublisher) *ScraperPublisherWorker {
	return &ScraperPublisherWorker{
		BaseWorker: bw,
		scraper:    scraper,
		publisher:  publisher,
	}
}

// Start begins the scraping process
func (w *ScraperPublisherWorker) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&w.Active, 0, 1) {
		return fmt.Errorf("worker %s is already running", w.Name())
	}

	for atomic.LoadInt32(&w.Active) == 1 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Perform single scrape operation
			articles, err := w.scraper.Scrape(ctx)
			if err != nil {
				log.Printf("Scraper error: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if len(articles) > 0 {
				if err := w.publisher.PublishArticles(ctx, articles); err != nil {
					log.Printf("Error publishing articles: %v", err)
				} else {
					log.Printf("Published %d articles", len(articles))
				}
			}

			// Wait for next context signal before scraping again
			return nil
		}
	}

	return nil
}

// Stop gracefully stops the worker
func (w *ScraperPublisherWorker) Stop() error {
	atomic.StoreInt32(&w.Active, 0)
	return nil
}

// Name returns the worker name
func (w *ScraperPublisherWorker) Name() string {
	return w.BaseWorker.Name
}
