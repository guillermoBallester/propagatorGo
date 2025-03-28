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
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration from the JSON file
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", *configPath, err)
	}

	log.Printf("Configuration loaded successfully from %s", *configPath)

	newsScraper := scraper.NewNewsScraper(&cfg.Scraper)
	newsScraper.Initialize()

	s := scheduler.NewScheduler(&cfg.Scheduler)

	initErr := s.Initialize()
	if initErr != nil {
		log.Fatalf("Failed to initialize scheduler: %v", initErr)
	}

	regErr := s.RegisterJobHandler("news-scraper", func(ctx context.Context) error {
		log.Println("Starting news scraping job...")
		_, scrapErr := newsScraper.Scrape(ctx)
		if scrapErr != nil {
			log.Printf("Scraping error: %v", scrapErr)
			return scrapErr
		}

		return nil
	})
	if regErr != nil {
		fmt.Printf("Failed to start news scrapper: %s", regErr)
	}

	s.Start()
	defer s.Stop()

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
