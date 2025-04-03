package repository

import (
	"context"
	"database/sql"
	"fmt"
	"propagatorGo/internal/database"
	"propagatorGo/internal/database/sqlc"
)

type ExtArticleRepo struct {
	queries *sqlc.Queries
	db      *sql.DB
}

// GetArticlesBySymbol retrieves articles for a specific stock symbol
// Returns articles, total count, and error
func (r *ExtArticleRepo) GetArticlesBySymbol(ctx context.Context, symbol string) ([]database.Article, int, error) {
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
func (r *ExtArticleRepo) GetArticlesBySite(ctx context.Context, siteName string) ([]database.Article, int, error) {
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

// GetLatestArticles retrieves the most recent articles across all sources and symbols
func (r *ExtArticleRepo) GetLatestArticles(ctx context.Context, limit int) ([]database.Article, error) {
	// Custom query since sqlc doesn't have it pre-generated
	query := `
		SELECT id, title, url, text, site_name, scraped_at, created_at, symbol 
		FROM articles 
		ORDER BY scraped_at DESC 
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying latest articles: %w", err)
	}
	defer rows.Close()

	var articles []database.Article
	for rows.Next() {
		var article database.Article
		var textNullable sql.NullString

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.URL,
			&textNullable,
			&article.SiteName,
			&article.ScrapedAt,
			&article.CreatedAt,
			&article.Symbol,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning article row: %w", err)
		}

		if textNullable.Valid {
			article.Text = textNullable.String
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating article rows: %w", err)
	}

	return articles, nil
}

// GetArticlesCount returns the total number of articles in the database
func (r *ExtArticleRepo) GetArticlesCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM articles`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting articles: %w", err)
	}
	return count, nil
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
