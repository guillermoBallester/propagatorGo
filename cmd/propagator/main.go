package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guillermoballester/propagatorGo/internal/config"
	"github.com/guillermoballester/propagatorGo/internal/constants"
	"github.com/guillermoballester/propagatorGo/internal/database"
	"github.com/guillermoballester/propagatorGo/internal/orchestrator"
	"github.com/guillermoballester/propagatorGo/internal/queue"
	"github.com/guillermoballester/propagatorGo/internal/repository"
	scraper "github.com/guillermoballester/propagatorGo/internal/scrapper"
	"github.com/guillermoballester/propagatorGo/internal/task"
	"github.com/guillermoballester/propagatorGo/internal/worker"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	cfg, errCfg := config.LoadConfig(*configPath)
	if errCfg != nil {
		log.Fatalf("Failed to load config: %v", errCfg)
	}

	dbClient, redisClient := initDB(cfg)
	deps := initWorkingDependencies(cfg, dbClient, redisClient)

	// Orchestrator
	o := orchestrator.NewOrchestrator(&cfg.Scheduler, deps)
	regErr := o.RegisterWorkerPool(config.WorkerConfig{
		PoolSize:   2,
		WorkerType: constants.WorkerTypeScraper,
		JobName:    constants.SourceYahoo + "-" + constants.WorkerTypeScraper,
		CronExpr:   "0 */30 * * * *",
		Source:     constants.SourceYahoo,
		Enabled:    true,
	})
	if regErr != nil {
		log.Panicf("Error registering Yahoo scraper pool: %v", regErr)
	}

	regErr = o.RegisterWorkerPool(config.WorkerConfig{
		PoolSize:   2,
		WorkerType: constants.WorkerTypeConsumer,
		JobName:    "writer" + constants.WorkerTypeConsumer,
		CronExpr:   "0 */30 * * * *",
		Enabled:    true,
	})

	o.Start()

	// Run initial jobs
	if err := o.RunJob(constants.SourceYahoo + "-" + constants.WorkerTypeScraper); err != nil {
		log.Printf("Failed to run news-scraper: %v", err)
	}

	go func() {
		log.Printf("Waiting 1 minute before running initial scraper job...")
		time.Sleep(1 * time.Minute)
		if err := o.RunJob("writer" + constants.WorkerTypeConsumer); err != nil {
			log.Printf("Failed to run news-scraper: %v", err)
		}
	}()

	// Wait for termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	o.Stop()
}

func initWorkingDependencies(cfg *config.Config, dbClient *database.PostgresClient, redisClient *queue.RedisClient) *orchestrator.WorkerDependencies {
	t := task.NewService(cfg, redisClient)
	r := repository.NewArticleRepository(dbClient.GetDB())
	s := scraper.NewScraperService(cfg, redisClient, t)
	f := worker.NewWorkerFactory(cfg, s, t, r)

	return &orchestrator.WorkerDependencies{
		ScraperSvc:    s,
		TaskService:   t,
		WorkerFactory: f,
	}
}

func initDB(cfg *config.Config) (*database.PostgresClient, *queue.RedisClient) {
	dbClient, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database client: %v", err)
	}
	//defer dbClient.Close()

	redisClient, err := queue.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	//defer redisClient.Close()

	return dbClient, redisClient
}
