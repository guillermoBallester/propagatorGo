package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Pool manages a collection of workers
type Pool struct {
	workers    []Worker
	numWorkers int
	wg         sync.WaitGroup
	mu         sync.Mutex
	isRunning  bool
}

// NewPool creates a new worker pool with the specified size
func NewPool(size int) *Pool {
	return &Pool{
		workers:    make([]Worker, 0, size),
		numWorkers: size,
		isRunning:  false,
	}
}

// AddWorker adds a worker to the pool
func (p *Pool) AddWorker(w Worker) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return fmt.Errorf("cannot add worker while pool is running")
	}

	if len(p.workers) >= p.numWorkers {
		return fmt.Errorf("worker pool is full (max: %d)", p.numWorkers)
	}

	p.workers = append(p.workers, w)
	return nil
}

// Start launches all workers in the pool
func (p *Pool) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.isRunning {
		p.mu.Unlock()
		return fmt.Errorf("worker pool is already running")
	}
	p.isRunning = true
	p.mu.Unlock()

	// Launch each worker in its own goroutine
	for _, w := range p.workers {
		p.wg.Add(1)
		worker := w // Capture variable for goroutine

		go func() {
			defer p.wg.Done()

			log.Printf("Starting worker: %s", worker.Name())
			if err := worker.Start(ctx); err != nil {
				log.Printf("Worker %s error: %v", worker.Name(), err)
			}
			log.Printf("Worker %s stopped", worker.Name())
		}()
	}

	return nil
}

// Stop gracefully shuts down all workers
func (p *Pool) Stop() {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	// Stop each worker
	for _, w := range p.workers {
		if err := w.Stop(); err != nil {
			log.Printf("Error stopping worker %s: %v", w.Name(), err)
		}
	}

	// Wait for all workers to complete
	p.wg.Wait()

	p.mu.Lock()
	p.isRunning = false
	p.mu.Unlock()
}

// Wait blocks until all workers have completed
func (p *Pool) Wait() {
	p.wg.Wait()
}
