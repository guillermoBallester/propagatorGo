package model

import "time"

// ArticleData represents the extracted data from an article
type ArticleData struct {
	Title     string
	URL       string
	Text      string
	SiteName  string
	Symbol    string    `json:"symbol"`
	ScrapedAt time.Time `json:"scraped_at"`
	CreatedAt time.Time `json:"created_at"`
}
