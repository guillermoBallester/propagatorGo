package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the main application configuration
type Config struct {
	App       AppConfig       `json:"app"`
	Scraper   ScraperConfig   `json:"scraper"`
	Scheduler SchedulerConfig `json:"scheduler"`
	Redis     RedisConfig     `json:"redis"`
	StockList StockList       `json:"stockList"`
	Database  DatabaseConfig  `json:"database"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslMode"`
}

// StockList represents the master list of stocks to be tracked
type StockList struct {
	Stocks []Stock `json:"stocks"`
}

// Stock defines a stock ticker to be tracked
type Stock struct {
	Symbol  string `json:"symbol"`
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled"`
}

// SiteConfig stores the selector configuration for each website
type SiteConfig struct {
	Name                 string   `json:"name"`
	URL                  string   `json:"url"`
	AllowedDomains       []string `json:"allowedDomains"`
	ArticleContainerPath string   `json:"articleContainerPath"`
	TitlePath            string   `json:"titlePath"`
	LinkPath             string   `json:"linkPath"`
	TextPath             string   `json:"textPath"`
	ImagePath            string   `json:"imagePath,omitempty"`
	Enabled              bool     `json:"enabled"`
}

type AppConfig struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Port      int    `json:"port"`
	LogLevel  string `json:"logLevel"`
	APIPrefix string `json:"apiPrefix"`
	Env       string `json:"env"` // development, production, testing
}

// ScraperConfig contains settings for web scraping
type ScraperConfig struct {
	UserAgent     string        `json:"userAgent"`
	MaxDepth      int           `json:"maxDepth"`
	MaxRetries    int           `json:"maxRetries"`
	RandomDelay   time.Duration `json:"randomDelay"`
	Sites         []SiteConfig  `json:"sites"`
	ParallelLimit int           `json:"parallelLimit"`
}

// SchedulerConfig contains settings for job scheduling
type SchedulerConfig struct {
	Jobs           []JobConfig   `json:"jobs"`
	DefaultTimeout time.Duration `json:"defaultTimeout"`
}

// JobConfig represents a schedulable job configuration
type JobConfig struct {
	Name        string        `json:"name"`
	CronExpr    string        `json:"cronExpr"`
	Timeout     time.Duration `json:"timeout"`
	RetryCount  int           `json:"retryCount"`
	Enabled     bool          `json:"enabled"`
	Description string        `json:"description"`
}

// RedisConfig represents Redis connection settings
type RedisConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

// WorkerConfig defines configuration for a worker pool
type WorkerConfig struct {
	PoolSize   int    `json:"poolSize"`
	WorkerType string `json:"workerType"`
	JobName    string `json:"jobName"`
	CronExpr   string `json:"cronExpr"`
	QueueName  string `json:"queueName,omitempty"`
	TaskType   string `json:"taskType,omitempty"`
	Source     string `json:"source,omitempty"`
	Enabled    bool   `json:"enabled"`
}

// LoadConfig loads the configuration from a JSON file
// Returns an error if the file doesn't exist or cannot be parsed
func LoadConfig(filePath string) (*Config, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig performs basic validation on the configuration
func validateConfig(cfg *Config) error {
	if cfg.App.Name == "" {
		return fmt.Errorf("app name is required")
	}

	if len(cfg.Scraper.Sites) == 0 {
		return fmt.Errorf("at least one scraper site must be configured")
	}

	hasEnabledSite := false
	for _, site := range cfg.Scraper.Sites {
		if site.Enabled {
			hasEnabledSite = true
			break
		}
	}
	if !hasEnabledSite {
		return fmt.Errorf("at least one scraper site must be enabled")
	}

	if len(cfg.Scheduler.Jobs) == 0 {
		return fmt.Errorf("at least one scheduler job must be configured")
	}

	if len(cfg.StockList.Stocks) == 0 {
		return fmt.Errorf("at least one stock list must be configured")
	}

	return nil
}
