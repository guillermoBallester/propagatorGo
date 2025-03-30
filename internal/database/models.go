package database

import "time"

// Article represents a news article in the database
type Article struct {
	ID        int64     `db:"id"`
	Title     string    `db:"title"`
	URL       string    `db:"url"`
	Text      string    `db:"text"`
	SiteName  string    `db:"site_name"`
	ScrapedAt time.Time `db:"scraped_at"`
	CreatedAt time.Time `db:"created_at"`
	Symbol    string    `db:"symbol"`

	// Relationships (not stored directly in the database)
	Tags []string `db:"-"`
}
