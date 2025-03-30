package scraper

import (
	"context"
	"fmt"
	"propagatorGo/internal/config"
	"propagatorGo/internal/queue"
	"sync"
)

// Service manages scrapers for different sources
type Service struct {
	config        *config.Config
	redisClient   *queue.RedisClient
	scrapersMutex sync.RWMutex
	scrapers      map[string]*NewsScraper
	publisher     *ArticlePublisher
}

// NewScraperService creates a new scraper service
func NewScraperService(cfg *config.Config, redis *queue.RedisClient) *Service {
	publisher := NewArticlePublisher(redis)
	return &Service{
		config:      cfg,
		redisClient: redis,
		scrapers:    make(map[string]*NewsScraper),
		publisher:   publisher,
	}
}

// GetPublisher returns the article publisher
func (s *Service) GetPublisher() *ArticlePublisher {
	return s.publisher
}

// ScrapeAndPublish performs both scraping and publishing in one operation
func (s *Service) ScrapeAndPublish(ctx context.Context, source string, symbol string) ([]ArticleData, error) {
	// Get the scraper for this source
	scraper, err := s.GetScraper(source)
	if err != nil {
		return nil, fmt.Errorf("error getting scraper: %w", err)
	}
	articles, err := scraper.Scrape(ctx, source, symbol)
	if err != nil {
		return nil, fmt.Errorf("error scraping: %w", err)
	}

	// If we have articles, publish them
	if len(articles) > 0 {
		if err := s.publisher.PublishArticles(ctx, articles); err != nil {
			return articles, fmt.Errorf("error publishing articles: %w", err)
		}
	}

	return articles, nil
}

// GetScraper returns (or creates) a scraper for a specific source
func (s *Service) GetScraper(source string) (*NewsScraper, error) {
	s.scrapersMutex.RLock()
	scraper, exists := s.scrapers[source]
	s.scrapersMutex.RUnlock()

	if exists {
		return scraper, nil
	}

	// Find the site config for this source
	var siteConfig *config.SiteConfig
	for _, site := range s.config.Scraper.Sites {
		if site.Name == source && site.Enabled {
			siteConfig = &site
			break
		}
	}

	if siteConfig == nil {
		return nil, fmt.Errorf("no configuration found for source: %s", source)
	}

	scraperConfig := &config.ScraperConfig{
		UserAgent:     s.config.Scraper.UserAgent,
		MaxDepth:      s.config.Scraper.MaxDepth,
		MaxRetries:    s.config.Scraper.MaxRetries,
		RandomDelay:   s.config.Scraper.RandomDelay,
		ParallelLimit: s.config.Scraper.ParallelLimit,
		Sites:         []config.SiteConfig{*siteConfig},
	}

	newScraper, err := NewNewsScraper(scraperConfig, siteConfig)
	if err != nil {
		return nil, err
	}

	s.scrapersMutex.Lock()
	s.scrapers[source] = newScraper
	s.scrapersMutex.Unlock()

	return newScraper, nil
}
