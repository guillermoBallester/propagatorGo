package scraper

import (
	"propagatorGo/internal/model"
	"testing"
	"time"
)

func TestSaveArticle(t *testing.T) {
	// Initialize scraper
	s := &NewsScraper{
		articles: make([]model.ArticleData, 0),
	}

	// Create test article
	testArticle := model.ArticleData{
		Title:     "Test Title",
		URL:       "https://example.com/test",
		Text:      "Test content",
		SiteName:  "TestSite",
		ScrapedAt: time.Now(),
	}

	s.saveArticle(testArticle)

	if len(s.articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(s.articles))
	}

	savedArticle := s.articles[0]
	if savedArticle.Title != testArticle.Title {
		t.Errorf("Expected title %s, got %s", testArticle.Title, savedArticle.Title)
	}
	if savedArticle.URL != testArticle.URL {
		t.Errorf("Expected URL %s, got %s", testArticle.URL, savedArticle.URL)
	}
}

func TestGetArticles(t *testing.T) {
	s := &NewsScraper{
		articles: make([]model.ArticleData, 0),
	}

	// Add test articles
	testArticles := []model.ArticleData{
		{
			Title:     "Article 1",
			URL:       "https://example.com/1",
			SiteName:  "TestSite",
			ScrapedAt: time.Now(),
		},
		{
			Title:     "Article 2",
			URL:       "https://example.com/2",
			SiteName:  "TestSite",
			ScrapedAt: time.Now(),
		},
	}

	for _, article := range testArticles {
		s.saveArticle(article)
	}

	articles := s.GetArticles()

	if len(articles) != len(testArticles) {
		t.Errorf("Expected %d articles, got %d", len(testArticles), len(articles))
	}

	origLen := len(s.articles)
	articles = append(articles, model.ArticleData{Title: "New Article"})
	if len(s.articles) != origLen {
		t.Error("GetArticles did not return a copy - original was modified")
	}
}

func TestResetArticles(t *testing.T) {
	s := &NewsScraper{
		articles: make([]model.ArticleData, 0),
	}

	testArticle := model.ArticleData{
		Title:     "Test Title",
		URL:       "https://example.com/test",
		SiteName:  "TestSite",
		ScrapedAt: time.Now(),
	}
	s.saveArticle(testArticle)

	if len(s.articles) != 1 {
		t.Fatalf("Failed to save article, got %d articles", len(s.articles))
	}

	s.resetArticles()

	if len(s.articles) != 0 {
		t.Errorf("Expected 0 articles after reset, got %d", len(s.articles))
	}
}
