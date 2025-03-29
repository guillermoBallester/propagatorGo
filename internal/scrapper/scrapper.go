package scraper

import (
	"context"
	"fmt"
	"log"
	"propagatorGo/internal/config"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

// SiteConfig stores the selector configuration for each website
type SiteConfig struct {
	Name                 string   `json:"name"`
	URL                  string   `json:"url"`
	AllowedDomains       []string `json:"allowedDomains"`
	ArticleContainerPath string   `json:"articleContainerPath"`
	TitlePath            string   `json:"titlePath"`
	LinkPath             string   `json:"linkPath"`
	TextPath             string   `json:"textPath"`
	ImagePath            string   `json:"imagePath,omitempty"`
	Enabled              bool     `json:"enabled"`
}

// NewsScraper is the main manager for scraping news
type NewsScraper struct {
	configs       []config.SiteConfig
	mainCollector *colly.Collector
	articles      []ArticleData
	articleMutex  sync.Mutex
}

// NewNewsScraper creates a new instance of the scraper
func NewNewsScraper(cfg *config.ScraperConfig) (*NewsScraper, error) {
	var allowedDomains []string
	for _, SiteConfig := range cfg.Sites {
		allowedDomains = append(allowedDomains, SiteConfig.AllowedDomains...)
	}

	col := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.UserAgent(cfg.UserAgent),
		colly.MaxDepth(cfg.MaxDepth),
		colly.AllowURLRevisit(),
	)

	err := col.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	col.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %s", r.Request.URL, err)
	})

	scraper := &NewsScraper{
		configs:       cfg.Sites,
		articles:      make([]ArticleData, 0),
		mainCollector: col,
	}

	return scraper, nil
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

		switch {
		case strings.Contains(strings.ToLower(siteConfig.Name), "yahoo"):
			collector.OnHTML(siteConfig.ArticleContainerPath, func(e *colly.HTMLElement) {
				if !strings.Contains(e.Attr("class"), "stream-items") {
					return
				}

				articles := s.extractYahooArticles(e, SiteConfig(siteConfig))
				for _, article := range articles {
					s.saveArticle(article)
					log.Printf("Scraped Yahoo article: %s", article.Title)
				}
			})
		}
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
		}(SiteConfig(config))
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
