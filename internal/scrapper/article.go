package scraper

import (
	"propagatorGo/internal/constants"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

// ArticleData represents the extracted data from an article
type ArticleData struct {
	Title     string
	URL       string
	Text      string
	SiteName  string
	ScrapedAt time.Time `json:"scraped_at"`
}

// saveArticle safely adds an article to the collection
func (s *NewsScraper) saveArticle(article ArticleData) {
	s.articleMutex.Lock()
	defer s.articleMutex.Unlock()

	s.articles = append(s.articles, article)
}

func (s *NewsScraper) extractYahooArticles(e *colly.HTMLElement) []ArticleData {
	var articles []ArticleData

	e.ForEach("li", func(_ int, li *colly.HTMLElement) {
		// Yahoo-specific extraction logic...
		liClass := li.Attr("class")
		if !strings.Contains(liClass, "stream-item") || !strings.Contains(liClass, "story-item") {
			return
		}

		article := ArticleData{
			SiteName:  constants.SourceYahoo,
			Title:     li.ChildText("h3"),
			URL:       li.ChildAttr("a", "href"),
			Text:      li.ChildText("p"),
			ScrapedAt: time.Now(),
		}

		if !strings.HasPrefix(article.URL, "http") {
			article.URL = e.Request.AbsoluteURL(article.URL)
		}

		if article.Title != "" && article.URL != "" {
			articles = append(articles, article)
		}
	})

	return articles
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
