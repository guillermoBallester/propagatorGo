package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"propagatorGo/internal/config"
	"propagatorGo/internal/scheduler"
	scraper "propagatorGo/internal/scrapper"
	"syscall"
)

func main() {

	cfg := loadConfig()
	scraperSrv, err := initServices(cfg)
	if err != nil {
		log.Fatal("Failed to initialize services")
	}

	s := scheduler.NewScheduler(&cfg.Scheduler)
	initErr := s.Initialize()
	if initErr != nil {
		log.Fatalf("Failed to initialize scheduler: %v", initErr)
	}

	s.Start()
	defer s.Stop()

	regErr := s.RegisterJobHandler("news-scraper", func(ctx context.Context) error {
		_, scrapErr := scraperSrv.Scrape(ctx)
		if scrapErr != nil {
			log.Printf("Scraping error: %v", scrapErr)
			return scrapErr
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

func initServices(cfg *config.Config) (*scraper.NewsScraper, error) {
	return scraper.NewNewsScraper(&cfg.Scraper)
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
