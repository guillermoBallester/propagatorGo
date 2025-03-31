package orchestrator

import (
	"context"
	"fmt"
	"log"
	"propagatorGo/internal/config"
	"propagatorGo/internal/scheduler"
	scraper "propagatorGo/internal/scrapper"
	"propagatorGo/internal/task"
	"propagatorGo/internal/worker"
	"time"
)

// Orchestrator manages worker pools and schedules their execution
type Orchestrator struct {
	scheduler *scheduler.Scheduler
	pools     map[string]*worker.Pool

	workerDeps *WorkerDependencies
}

// WorkerDependencies contains all dependencies needed for various worker types
type WorkerDependencies struct {
	ScraperSvc    *scraper.Service
	TaskService   *task.Service
	WorkerFactory *worker.Factory
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
func (o *Orchestrator) RegisterWorkerPool(cfg config.WorkerConfig) error {
	pool := worker.NewPool(cfg.PoolSize)

	err := o.workerDeps.TaskService.EnqueueAll(context.Background(), cfg.TaskType, cfg.Source)
	if err != nil {
		return err
	}

	// Create and add workers based on type
	for i := 0; i < cfg.PoolSize; i++ {
		w, workerErr := o.workerDeps.WorkerFactory.CreateWorker(i, cfg.WorkerType)
		if workerErr != nil {
			return fmt.Errorf("error creating worker: %w", workerErr)
		}
		if poolErr := pool.AddWorker(w); poolErr != nil {
			return fmt.Errorf("error adding worker: %w", poolErr)
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
