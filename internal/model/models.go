package model

// ArticleData represents the extracted data from an article
type ArticleData struct {
	Title     string
	URL       string
	Text      string
	SiteName  string
	Symbol    string `json:"symbol"`
	ScrapedAt string `json:"scraped_at"`
	CreatedAt string `json:"created_at"`
}
