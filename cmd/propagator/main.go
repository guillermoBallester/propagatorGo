package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"propagatorGo/internal/config"
	"propagatorGo/internal/orchestrator"
	"propagatorGo/internal/queue"
	scraper "propagatorGo/internal/scrapper"
	"propagatorGo/internal/worker"
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

	deps := &orchestrator.WorkerDependencies{
		Scraper:     handler.Scraper,
		Publisher:   handler.Publisher,
		RedisClient: handler.Redis,
		//DBClient:    dbClient,
	}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(&cfg.Scheduler, deps)

	// Register worker pools
	regErr := orch.RegisterWorkerPool(orchestrator.WorkerConfig{
		PoolSize:    2,
		WorkerType:  worker.ScraperPublisherType,
		JobName:     "news-scraper",
		CronExpr:    "0 */5 * * * *", // Every 5 minutes
		Description: "Scrapes news articles",
		Enabled:     true,
	})
	if regErr != nil {
		log.Panicf("Error registering pool: %v", regErr)
	}

	// Start the orchestrator
	orch.Start()

	// Run initial jobs
	if err := orch.RunJob("news-scraper"); err != nil {
		log.Printf("Failed to run news-scraper: %v", err)
	}

	// Wait for termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	orch.Stop()
}

func initServices(cfg *config.Config) (*SrvHandler, error) {
	clientRedis, err := queue.NewRedisClient(&cfg.Redis)
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
