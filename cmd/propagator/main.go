package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"propagatorGo/internal/config"
	"propagatorGo/internal/queue"
	"propagatorGo/internal/scheduler"
	scraper "propagatorGo/internal/scrapper"
	"syscall"
)

type SrvHandler struct {
	Scraper   *scraper.NewsScraper
	Redis     *queue.RedisClient
	Publisher *scraper.ArticlePublisher
}

func main() {

	cfg := loadConfig()
	handler, err := initServices(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services, %v", err)
	}

	defer func(redisClient *queue.RedisClient) {
		err := handler.Redis.Close()
		if err != nil {

		}
	}(handler.Redis)

	s := scheduler.NewScheduler(&cfg.Scheduler)
	initErr := s.Initialize()
	if initErr != nil {
		log.Fatalf("Failed to initialize scheduler: %v", initErr)
	}

	s.Start()
	defer s.Stop()

	regErr := s.RegisterJobHandler("news-scraper", func(ctx context.Context) error {
		articles, scrapErr := handler.Scraper.Scrape(ctx)
		if scrapErr != nil {
			log.Printf("Scraping error: %v", scrapErr)
			return scrapErr
		}

		if len(articles) > 0 {
			err := handler.Publisher.PublishArticles(ctx, articles)
			if err != nil {
				return fmt.Errorf("error while publishing articles: %w", err)
			}
		}

		return nil
	})
	if regErr != nil {
		fmt.Printf("Failed to start news scrapper: %s", regErr)
	}

	runErr := s.RunJob("news-scraper")
	if runErr != nil {
		return
	}

	// Keep the program running
	log.Println("Scheduler running. Press Ctrl+C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func initServices(cfg *config.Config) (*SrvHandler, error) {
	clientRedis, err := queue.NewRedisClient(cfg.Redis.Address, cfg.Redis.Password, 0)
	if err != nil {
		return nil, err
	}

	publisher := scraper.NewArticlePublisher(clientRedis)

	svcScraper, err := scraper.NewNewsScraper(&cfg.Scraper)
	if err != nil {
		return nil, err
	}

	return &SrvHandler{
		Scraper:   svcScraper,
		Redis:     clientRedis,
		Publisher: publisher,
	}, nil
}

func loadConfig() *config.Config {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", *configPath, err)
	}

	log.Printf("Configuration loaded successfully from %s", *configPath)

	return cfg
}
