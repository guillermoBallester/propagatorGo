package scraper

import (
	"strings"
	"time"

	"github.com/gocolly/colly"
)

// ArticleData represents the extracted data from an article
type ArticleData struct {
	Title     string
	URL       string
	Text      string
	ImageURL  string
	SiteName  string
	ScrapedAt time.Time `json:"scraped_at"`
}

// saveArticle safely adds an article to the collection
func (s *NewsScraper) saveArticle(article ArticleData) {
	s.articleMutex.Lock()
	defer s.articleMutex.Unlock()

	s.articles = append(s.articles, article)
}

// extractArticle parses HTML elements to extract article data
func (s *NewsScraper) extractArticle(e *colly.HTMLElement, config SiteConfig) ArticleData {
	article := ArticleData{
		SiteName:  config.Name,
		ScrapedAt: time.Now(),
	}

	if config.TitlePath != "" {
		e.ForEach(config.TitlePath, func(_ int, el *colly.HTMLElement) {
			if article.Title == "" {
				titleText := el.Text
				titleText = strings.TrimPrefix(titleText, "::before")
				titleText = strings.TrimSpace(titleText)
				article.Title = titleText
			}
		})
	}

	if config.LinkPath != "" {
		e.ForEach(config.LinkPath, func(_ int, el *colly.HTMLElement) {
			if article.URL == "" {
				href := el.Attr("href")
				if href != "" {
					article.URL = e.Request.AbsoluteURL(href)
				}
			}
		})
	}

	if config.TextPath != "" {
		article.Text = strings.TrimSpace(e.DOM.Parent().Find(config.TextPath).Text())
	}

	return article
}

// GetArticles returns the collected articles
func (s *NewsScraper) GetArticles() []ArticleData {
	s.articleMutex.Lock()
	defer s.articleMutex.Unlock()

	// Return a copy to avoid race conditions
	articlesCopy := make([]ArticleData, len(s.articles))
	copy(articlesCopy, s.articles)

	return articlesCopy
}

// resetArticles clears the articles slice
func (s *NewsScraper) resetArticles() {
	s.articleMutex.Lock()
	defer s.articleMutex.Unlock()

	s.articles = make([]ArticleData, 0)
}
