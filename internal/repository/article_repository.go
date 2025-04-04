// internal/repository/article_repository.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"propagatorGo/internal/database"
	"propagatorGo/internal/database/sqlc"
)

// ArticleRepository handles database operations for articles
type ArticleRepository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

// NewArticleRepository creates a new article repository
func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{
		queries: sqlc.New(db),
		db:      db,
	}
}

// SaveArticle saves or updates an article in the database
func (r *ArticleRepository) SaveArticle(ctx context.Context, article database.Article) error {
	params := sqlc.CreateArticleParams{
		Title:     article.Title,
		Url:       article.URL,
		Text:      sql.NullString{String: article.Text, Valid: article.Text != ""},
		SiteName:  article.SiteName,
		ScrapedAt: article.ScrapedAt,
		Symbol:    article.Symbol,
	}

	_, err := r.queries.CreateArticle(ctx, params)
	return err
}

// GetArticlesBySymbol retrieves articles for a specific stock symbol
// Returns articles, total count, and error
func (r *ArticleRepository) GetArticlesBySymbol(ctx context.Context, symbol string) ([]database.Article, int, error) {
	dbArticles, err := r.queries.GetArticleBySymbol(ctx, symbol)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	countQuery := `SELECT COUNT(*) FROM articles WHERE symbol = $1`
	var total int
	err = r.db.QueryRowContext(ctx, countQuery, symbol).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting articles: %w", err)
	}

	return mapSQLCArticlesToModels(dbArticles), total, nil
}

// GetArticlesBySite retrieves articles from a specific site
// Returns articles, total count, and error
func (r *ArticleRepository) GetArticlesBySite(ctx context.Context, siteName string) ([]database.Article, int, error) {
	dbArticles, err := r.queries.GetArticleBySite(ctx, siteName)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	countQuery := `SELECT COUNT(*) FROM articles WHERE site_name = $1`
	var total int
	err = r.db.QueryRowContext(ctx, countQuery, siteName).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting articles: %w", err)
	}

	return mapSQLCArticlesToModels(dbArticles), total, nil
}

// Helper function to map sqlc Article model to our domain model
func mapSQLCArticleToModel(dbArticle sqlc.Article) database.Article {
	return database.Article{
		ID:        int64(dbArticle.ID),
		Title:     dbArticle.Title,
		URL:       dbArticle.Url,
		Text:      dbArticle.Text.String,
		SiteName:  dbArticle.SiteName,
		ScrapedAt: dbArticle.ScrapedAt,
		CreatedAt: dbArticle.CreatedAt,
		Symbol:    dbArticle.Symbol,
	}
}

// Helper function to map a slice of sqlc Articles to our domain models
func mapSQLCArticlesToModels(dbArticles []sqlc.Article) []database.Article {
	articles := make([]database.Article, len(dbArticles))
	for i, dbArticle := range dbArticles {
		articles[i] = mapSQLCArticleToModel(dbArticle)
	}
	return articles
}
