package orchestrator

import (
	"context"
	"fmt"
	"log"
	"propagatorGo/internal/config"
	"propagatorGo/internal/database"
	"propagatorGo/internal/queue"
	"propagatorGo/internal/scheduler"
	scraper "propagatorGo/internal/scrapper"
	"propagatorGo/internal/stock"
	"propagatorGo/internal/worker"
	"time"
)

// WorkerConfig defines configuration for a worker pool
type WorkerConfig struct {
	PoolSize    int    `json:"poolSize"`
	WorkerType  string `json:"workerType"`
	JobName     string `json:"jobName"`
	CronExpr    string `json:"cronExpr"`
	QueueName   string `json:"queueName,omitempty"`
	Source      string `json:"source,omitempty"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// Orchestrator manages worker pools and schedules their execution
type Orchestrator struct {
	scheduler  *scheduler.Scheduler
	pools      map[string]*worker.Pool
	workerDeps *WorkerDependencies
}

// WorkerDependencies contains all dependencies needed for various worker types
type WorkerDependencies struct {
	ScraperSvc  *scraper.Service
	Publisher   *scraper.ArticlePublisher //TODO Publisher SHOULD be part of the scrapper Svc
	RedisClient *queue.RedisClient
	StockSvc    *stock.Service
	DBClient    *database.PostgresClient
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(schedulerCfg *config.SchedulerConfig, deps *WorkerDependencies) *Orchestrator {
	return &Orchestrator{
		scheduler:  scheduler.NewScheduler(schedulerCfg),
		pools:      make(map[string]*worker.Pool),
		workerDeps: deps,
	}
}

// RegisterWorkerPool creates and registers a worker pool
func (o *Orchestrator) RegisterWorkerPool(cfg WorkerConfig) error {
	pool := worker.NewPool(cfg.PoolSize)

	// Create and add workers based on type
	for i := 0; i < cfg.PoolSize; i++ {
		var w worker.Worker

		switch cfg.WorkerType {
		case worker.ScraperPublisherType:
			err := o.workerDeps.StockSvc.EnqueueAllStocks(context.Background(), stock.TaskTypeScrape, "yahoo")
			if err != nil {
				return err
			}
			scrapWorker := worker.NewBaseWorker(i, fmt.Sprintf("%s%d", worker.ScraperPublisherType, i), cfg.WorkerType)
			w = worker.NewScraperWorker(
				scrapWorker,
				o.workerDeps.ScraperSvc,
				o.workerDeps.StockSvc,
				"yahoo")
		case worker.ConsumerWriterType:
		case "api":
			// Add implementation for API worker
		default:
			return fmt.Errorf("unknown worker type: %s", cfg.WorkerType)
		}

		if err := pool.AddWorker(w); err != nil {
			return fmt.Errorf("error adding worker: %w", err)
		}
	}

	o.pools[cfg.JobName] = pool
	if err := o.registerJobHandler(cfg.JobName, cfg.CronExpr, pool); err != nil {
		return fmt.Errorf("error registering job: %w", err)
	}

	return nil
}

// registerJobHandler adds a job to the scheduler for a worker pool
func (o *Orchestrator) registerJobHandler(name string, cronExpr string, pool *worker.Pool) error {
	return o.scheduler.AddJob(name, cronExpr, 5*time.Minute, func(ctx context.Context) error {
		log.Printf("Starting worker pool for job: %s", name)

		// Create a context that can be cancelled
		poolCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Start the worker pool
		if err := pool.Start(poolCtx); err != nil {
			return fmt.Errorf("error starting worker pool: %w", err)
		}

		// Wait for the pool to finish or context to be cancelled
		select {
		case <-ctx.Done():
			log.Printf("Job %s cancelled", name)
		case <-time.After(4 * time.Minute): // Leave some buffer before timeout
			log.Printf("Job %s maximum runtime reached", name)
		}

		// Stop the pool
		pool.Stop()
		pool.Wait()

		log.Printf("Worker pool for job %s completed", name)
		return nil
	})
}

// Start starts the orchestrator
func (o *Orchestrator) Start() {
	o.scheduler.Start()
	log.Println("Orchestrator started")
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	o.scheduler.Stop()
	log.Println("Orchestrator stopped")
}

// RunJob immediately runs a job
func (o *Orchestrator) RunJob(name string) error {
	log.Printf("Starting job: %s", name)
	return o.scheduler.RunJob(name)
}
