package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ScraperPublisherType = "ScrapperPublisher"
	ConsumerWriterType   = "ConsumerWriter"
)

// BaseWorker contains common attributes for all worker types
type BaseWorker struct {
	ID           int       // Unique identifier for the worker
	workerName   string    // Human-readable name
	typeOfWorker string    // Type of worker (e.g., "scraper", "consumer")
	active       int32     // Atomic flag for tracking active state
	StartTime    time.Time // When the worker started
	Stats        *Stats    // Performance statistics
}

// Stats tracks operational metrics for a worker
type Stats struct {
	ItemsProcessed  int64         // Total items processed
	ErrorCount      int64         // Number of errors encountered
	LastProcessed   time.Time     // When last item was processed
	ProcessingTime  time.Duration // Total time spent processing
	AverageItemTime time.Duration // Average time per item
	mu              sync.Mutex    // Mutex for updating stats
}

// Worker represents a generic worker that can process tasks
type Worker interface {
	// Start begins processing tasks
	Start(ctx context.Context) error

	// Stop gracefully shuts down the worker
	Stop() error

	// Name returns the worker's name
	Name() string
}

// Stop gracefully stops the worker by setting the Active flag to 0
func (w *BaseWorker) Stop() error {
	atomic.StoreInt32(&w.active, 0)
	return nil
}

// Name returns the worker's name
func (w *BaseWorker) Name() string {
	return w.workerName
}

// IsActive checks if the worker is currently active
func (w *BaseWorker) IsActive() bool {
	return atomic.LoadInt32(&w.active) == 1
}

// SetActive sets the active state of the worker
// Returns true if the state was changed, false if it was already in the desired state
func (w *BaseWorker) SetActive(active bool) bool {
	if active {
		return atomic.CompareAndSwapInt32(&w.active, 0, 1)
	}
	return atomic.CompareAndSwapInt32(&w.active, 1, 0)
}

// NewBaseWorker creates a new base worker with the given parameters
func NewBaseWorker(id int, name, workerType string) BaseWorker {
	return BaseWorker{
		ID:           id,
		workerName:   name,
		typeOfWorker: workerType,
		active:       0,
		StartTime:    time.Time{},
		Stats:        &Stats{},
	}
}
