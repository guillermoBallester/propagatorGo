package worker

import (
	"context"
	"sync"
	"time"
)

// BaseWorker contains common attributes for all worker types
type BaseWorker struct {
	ID        int       // Unique identifier for the worker
	Name      string    // Human-readable name
	Type      string    // Type of worker (e.g., "scraper", "consumer")
	Active    int32     // Atomic flag for tracking active state
	StartTime time.Time // When the worker started
	Stats     *Stats    // Performance statistics
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
