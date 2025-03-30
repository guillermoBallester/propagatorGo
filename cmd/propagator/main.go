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
	"propagatorGo/internal/stock"
	"propagatorGo/internal/worker"
	"syscall"
)

func main() {

	configPath := flag.String("config", "config.json", "Path to configuration file")
	cfg, errCfg := config.LoadConfig(*configPath)
	if errCfg != nil {
		log.Fatalf("Failed to load config: %v", errCfg)
	}

	redisClient, err := queue.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	defer redisClient.Close()

	deps := &orchestrator.WorkerDependencies{
		ScraperSvc:  scraper.NewScraperService(cfg, redisClient),
		RedisClient: redisClient,
		StockSvc:    stock.NewService(cfg, redisClient),
	}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(&cfg.Scheduler, deps)

	regErr := orch.RegisterWorkerPool(orchestrator.WorkerConfig{
		PoolSize:    4,
		WorkerType:  worker.ScraperPublisherType,
		JobName:     "yahoo-scraper",
		CronExpr:    "0 */30 * * * *",
		Source:      "yahoo",
		Description: "Scrapes Yahoo Finance for stock news",
		Enabled:     true,
	})
	if regErr != nil {
		log.Panicf("Error registering Yahoo scraper pool: %v", regErr)
	}

	// Start the orchestrator
	orch.Start()

	// Run initial jobs
	if err := orch.RunJob("yahoo-scraper"); err != nil {
		log.Printf("Failed to run news-scraper: %v", err)
	}

	// Wait for termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	orch.Stop()
}
