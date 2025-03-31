// internal/repository/article_repository.go
package repository

import (
	"context"
	"database/sql"
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

// GetArticleByURL retrieves an article by its URL
func (r *ArticleRepository) GetArticleByURL(ctx context.Context, url string) (database.Article, error) {
	dbArticle, err := r.queries.GetArticleByURL(ctx, url)
	if err != nil {
		return database.Article{}, err
	}

	return database.Article{
		ID:        int64(dbArticle.ID),
		Title:     dbArticle.Title,
		URL:       dbArticle.Url,
		Text:      dbArticle.Text.String,
		SiteName:  dbArticle.SiteName,
		ScrapedAt: dbArticle.ScrapedAt,
		CreatedAt: dbArticle.CreatedAt,
		Symbol:    dbArticle.Symbol,
	}, nil
}

// GetArticlesBySymbol retrieves articles for a specific stock symbol
func (r *ArticleRepository) GetArticlesBySymbol(ctx context.Context, symbol string) ([]database.Article, error) {
	dbArticles, err := r.queries.GetArticleBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}

	articles := make([]database.Article, len(dbArticles))
	for i, dbArticle := range dbArticles {
		articles[i] = database.Article{
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

	return articles, nil
}
