package model

import "time"

// ArticleData represents the extracted data from an article
type ArticleData struct {
	Title     string
	URL       string
	Text      string
	SiteName  string
	ScrapedAt time.Time `json:"scraped_at"`
}
