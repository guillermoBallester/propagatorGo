package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"propagatorGo/internal/config"
	"testing"
	"time"

	"github.com/gocolly/colly"
)

// setupTestServer creates a test HTTP server for scraping
func setupTestServer() *httptest.Server {
	// Mock HTML response
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Site</title>
	</head>
	<body>
		<header class="segmento-header">
			<h2 class="segmento-title"><a href="/article/1">Test Article 1</a></h2>
			<p class="segmento-lead">This is the first test article.</p>
		</header>
		<header class="segmento-header">
			<h2 class="segmento-title"><a href="/article/2">Test Article 2</a></h2>
			<p class="segmento-lead">This is the second test article.</p>
		</header>
	</body>
	</html>
	`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))

	return server
}

func initConfig() *config.ScraperConfig {
	return &config.ScraperConfig{
		UserAgent:     "Mozilla/5.0 (Test User Agent)",
		MaxDepth:      2,
		MaxRetries:    3,
		RandomDelay:   5 * time.Second,
		ParallelLimit: 2,
		Sites: []config.SiteConfig{
			{
				Name:                 "TestSite",
				URL:                  "https://example.com",
				AllowedDomains:       []string{"example.com"},
				ArticleContainerPath: "header.segmento-header",
				TitlePath:            "h2.segmento-title a",
				LinkPath:             "h2.segmento-title a",
				TextPath:             "p.segmento-lead",
				Enabled:              true,
			},
		},
	}
}

func TestNewNewsScraper(t *testing.T) {

	configs := initConfig()
	scraper, initErr := NewNewsScraper(configs)
	if initErr != nil {
		t.Fatalf("Error initializing scraper: %v", initErr)
	}

	if len(scraper.configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(scraper.configs))
	}

	if scraper.configs[0].Name != "TestSite" {
		t.Errorf("Expected config name 'TestSite', got '%s'", scraper.configs[0].Name)
	}

	if scraper.mainCollector == nil {
		t.Error("Expected collector to be initialized")
	}
}

func TestScrape(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	configs := initConfig()
	scraper, initErr := NewNewsScraper(configs)
	if initErr != nil {
		t.Fatalf("Error initializing scraper: %v", initErr)
	}

	scraper.mainCollector = colly.NewCollector()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	articles, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}
	if len(articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles))
	}
	for i, article := range articles {
		expectedTitle := ""
		expectedText := ""

		switch i {
		case 0:
			expectedTitle = "Test Article 1"
			expectedText = "This is the first test article."
		case 1:
			expectedTitle = "Test Article 2"
			expectedText = "This is the second test article."
		}

		if article.Title != expectedTitle {
			t.Errorf("Article %d: Expected title '%s', got '%s'", i, expectedTitle, article.Title)
		}

		if article.Text != expectedText {
			t.Errorf("Article %d: Expected text '%s', got '%s'", i, expectedText, article.Text)
		}

		if article.SiteName != "TestSite" {
			t.Errorf("Article %d: Expected site name 'TestSite', got '%s'", i, article.SiteName)
		}
	}
}

func TestRegisterHTMLHandlers(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	configs := initConfig()
	scraper, initErr := NewNewsScraper(configs)
	if initErr != nil {
		t.Fatalf("Error initializing scraper: %v", initErr)
	}
	scraper.mainCollector = colly.NewCollector()

	ctx := context.Background()
	contextCollector := scraper.createContextCollector(ctx)
	scraper.registerHTMLHandlers(contextCollector)
	contextCollector.Visit(server.URL)
	contextCollector.Wait()

	articles := scraper.GetArticles()
	if len(articles) != 2 {
		t.Errorf("Expected 2 articles after handlers registered and site visited, got %d", len(articles))
	}
}

func TestStartScraping(t *testing.T) {
	// Set up test server
	server := setupTestServer()
	defer server.Close()
	configs := initConfig()
	scraper, initErr := NewNewsScraper(configs)
	if initErr != nil {
		t.Fatalf("Error initializing scraper: %v", initErr)
	}
	scraper.mainCollector = colly.NewCollector()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	contextCollector := scraper.createContextCollector(ctx)
	scraper.registerHTMLHandlers(contextCollector)
	done, errChan := scraper.startScraping(ctx, contextCollector)

	err := scraper.waitForCompletion(ctx, done, errChan, contextCollector)
	if err != nil {
		t.Fatalf("Scraping failed: %v", err)
	}
	articles := scraper.GetArticles()
	if len(articles) != 2 {
		t.Errorf("Expected 2 articles after scraping, got %d", len(articles))
	}
}
