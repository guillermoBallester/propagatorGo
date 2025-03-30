// scraper_worker.go
package worker

import (
	"context"
	"fmt"
	"log"
	scraper "propagatorGo/internal/scrapper"
	"propagatorGo/internal/stock"
	"time"
)

// ScraperPublisherWorker scrapes websites and publishes to Redis
type ScraperPublisherWorker struct {
	BaseWorker
	scraperService *scraper.Service
	stockService   *stock.Service
	source         string
}

// NewScraperWorker creates a new scraper worker
func NewScraperWorker(bw BaseWorker, scraperSvc *scraper.Service, stockSvc *stock.Service, source string) *ScraperPublisherWorker {
	return &ScraperPublisherWorker{
		BaseWorker:     bw,
		scraperService: scraperSvc,
		stockService:   stockSvc,
		source:         source,
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
			/*queueLength, err := w.stockService.QueueLength(ctx, stock.TaskTypeScrape, w.source)
			if err != nil {
				log.Printf("Error checking queue length: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// If queue is empty, we're done
			if queueLength == 0 {
				log.Printf("All tasks completed for %s", w.source)
				return nil
			}*/ //TODO Check queue length
			w.Stats.RecordStart()

			task, err := w.stockService.GetNextTask(ctx, stock.TaskTypeScrape, 5)
			if err != nil {
				log.Printf("Error getting stock task: %v", err)
				w.Stats.RecordItemFailed()
				time.Sleep(1 * time.Second)
				continue
			}

			// If no task returned within timeout, try again
			if task == nil {
				continue
			}

			// Process the stock task
			log.Printf("Worker %s processing stock: %s from source %s",
				w.Name(), task.Stock.Symbol, w.source)

			articles, err := w.scraperService.ScrapeAndPublish(ctx, w.source, task.Stock.Symbol)
			if err != nil {
				log.Printf("Error processing stock %s: %v", task.Stock.Symbol, err)
				w.Stats.RecordItemFailed()
				continue
			}

			w.Stats.RecordItemProcessed()
			stats := w.Stats.GetSnapshot()
			log.Printf("[%s] Task completed for %s. Articles: %d, Total processed: %d, Successful: %d, Failed: %d",
				w.Name(),
				task.Stock.Symbol,
				len(articles),
				stats.ItemsProcessed,
				stats.ItemsSuccessful,
				stats.ItemsFailed)
		}
	}

	return nil
}

// processScrapeTask handles the actual scraping based on the source
func (w *ScraperPublisherWorker) processScrapeTask(ctx context.Context, task *stock.Task) ([]scraper.ArticleData, error) {
	switch task.Source {
	case "yahoo":
		return w.scrapeYahoo(ctx, task.Stock.Symbol)
	default:
		return nil, fmt.Errorf("unsupported source: %s", task.Source)
	}
}

// scrapeYahooFinance scrapes Yahoo Finance for a specific stock
func (w *ScraperPublisherWorker) scrapeYahoo(ctx context.Context, symbol string) ([]scraper.ArticleData, error) {
	yahooScrapper, getErr := w.scraperService.GetScraper("yahoo")
	if getErr != nil {
		return nil, getErr //TODO define better error.
	}

	return yahooScrapper.Scrape(ctx, "yahoo", symbol)
}
