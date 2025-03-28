package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"propagatorGo/internal/scheduler"
	scraper "propagatorGo/internal/scrapper"
	"syscall"
	"time"
)

func main() {

	configs := []scraper.SiteConfig{
		{
			Name:                 "OkDiario",
			URL:                  "https://okdiario.com",
			AllowedDomains:       []string{"okdiario.com", "www.okdiario.com"},
			ArticleContainerPath: "header.segmento-header",
			TitlePath:            "h2.segmento-title a",
			LinkPath:             "h2.segmento-title a",
			TextPath:             "p.segmento-lead",
		},
	}

	newsScraper := scraper.NewNewsScraper(configs)
	newsScraper.Initialize()

	s := scheduler.NewScheduler()
	s.Start()
	defer s.Stop()

	err := s.AddJob("news-scraper", "0 */5 * * * *", 3*time.Minute, func(ctx context.Context) error {
		log.Println("Starting news scraping job...")

		_, err := newsScraper.Scrape(ctx)
		if err != nil {
			log.Printf("Scraping error: %v", err)
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Failed to start news scrapper: %s", err)
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
