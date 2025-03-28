package scraper

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

// SiteConfig stores the selector configuration for each website
type SiteConfig struct {
	Name                 string   // Name of the website
	URL                  string   // Main URL of the website
	AllowedDomains       []string // Allowed domains for scraping
	ArticleContainerPath string   // Selector for article containers
	TitlePath            string   // Selector for article titles
	LinkPath             string   // Selector for article links
	TextPath             string   // Selector for article text/summary
	ImagePath            string   // Selector for article images (optional)
}

// ArticleData represents the extracted data from an article
type ArticleData struct {
	Title     string
	URL       string
	Text      string
	ImageURL  string
	SiteName  string
	ScrapedAt time.Time `json:"scraped_at"`
}

// NewsScraper is the main manager for scraping news
type NewsScraper struct {
	configs       []SiteConfig
	mainCollector *colly.Collector
	articles      []ArticleData
	articleMutex  sync.Mutex
}

// NewNewsScraper creates a new instance of the scraper
func NewNewsScraper(configs []SiteConfig) *NewsScraper {
	var allowedDomains []string
	for _, config := range configs {
		allowedDomains = append(allowedDomains, config.AllowedDomains...)
	}

	col := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
		colly.MaxDepth(2),
	)

	scraper := &NewsScraper{
		configs:       configs,
		articles:      make([]ArticleData, 0),
		mainCollector: col,
	}

	return scraper
}

// Initialize sets up the collectors and prepares them for scraping
func (s *NewsScraper) Initialize() {
	s.mainCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 10 * time.Second,
	})

	s.mainCollector.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %s", r.Request.URL, err)
	})
}

// Scrape extracts information from an article preview
func (s *NewsScraper) Scrape(ctx context.Context) ([]ArticleData, error) {
	s.resetArticles()

	ctxCollector := s.createContextCollector(ctx)

	s.registerHTMLHandlers(ctxCollector)

	done, errChan := s.startScraping(ctx, ctxCollector)

	err := s.waitForCompletion(ctx, done, errChan, ctxCollector)
	if err != nil {
		return s.GetArticles(), err
	}

	return s.GetArticles(), nil
}

// Start begins the scraping process for all configurations
func (s *NewsScraper) Start() {
	for _, config := range s.configs {
		fmt.Printf("Starting to scrape: %s\n", config.Name)
		err := s.mainCollector.Visit(config.URL)
		if err != nil {
			log.Printf("Error visiting %s: %s", config.URL, err)
		}
	}
	s.mainCollector.Wait()
}

// createContextCollector creates a collector with context cancellation
func (s *NewsScraper) createContextCollector(ctx context.Context) *colly.Collector {
	ctxCollector := s.mainCollector.Clone()

	// Set up context cancellation
	ctxCollector.OnRequest(func(r *colly.Request) {
		select {
		case <-ctx.Done():
			r.Abort()
		default:
			// Continue with request
		}
	})

	return ctxCollector
}

// registerHTMLHandlers sets up HTML handlers for article extraction
func (s *NewsScraper) registerHTMLHandlers(collector *colly.Collector) {
	for _, config := range s.configs {
		siteConfig := config

		collector.OnHTML(siteConfig.ArticleContainerPath, func(e *colly.HTMLElement) {
			article := s.extractArticle(e, siteConfig)

			if article.Title != "" && article.URL != "" {
				s.saveArticle(article)
				log.Printf("Scraped article: %s from %s", article.Title, siteConfig.Name)
			}
		})
	}
}

// startScraping begins the scraping process for all sites
func (s *NewsScraper) startScraping(ctx context.Context, collector *colly.Collector) (chan bool, chan error) {
	done := make(chan bool)
	errChan := make(chan error, len(s.configs))

	var wg sync.WaitGroup

	// Start visiting URLs for each config
	for _, config := range s.configs {
		wg.Add(1)
		go func(cfg SiteConfig) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				// Continue with scraping
			}

			log.Printf("Starting to scrape: %s from %s", cfg.Name, cfg.URL)
			err := collector.Visit(cfg.URL)
			if err != nil {
				log.Printf("Error visiting %s: %v", cfg.URL, err)
				errChan <- fmt.Errorf("error scraping %s: %w", cfg.Name, err)
			}
		}(config)
	}

	// goroutine that manages completion signal
	go func() {
		wgDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(wgDone)
		}()

		select {
		case <-wgDone:
		case <-ctx.Done():
			<-wgDone
		}
		close(done)
	}()

	return done, errChan
}

// waitForCompletion waits for scraping to complete or context to cancel
func (s *NewsScraper) waitForCompletion(ctx context.Context, done chan bool, errChan chan error, collector *colly.Collector) error {
	select {
	case <-done:
		collector.Wait() //returns when collector colly is finished
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-errChan:
		return err
	default:
		// No errors
	}

	return nil
}
