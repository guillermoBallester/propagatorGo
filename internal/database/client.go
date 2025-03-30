package database

import (
	"database/sql"
	"fmt"
	"propagatorGo/internal/database/sqlc"

	_ "github.com/lib/pq"
)

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

// PostgresClient handles database operations
type PostgresClient struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// New creates a new database client
func New(cfg Config) (*PostgresClient, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to PostgreSQL: %w", err)
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	queries := sqlc.New(db)

	return &PostgresClient{
		db:      db,
		queries: queries,
	}, nil
}

// Close closes the database connection
func (c *PostgresClient) Close() error {
	return c.db.Close()
}
