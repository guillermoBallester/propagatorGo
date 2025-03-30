package worker

import (
	"sync"
	"sync/atomic"
	"time"
)

// Stats tracks operational metrics for a worker
type Stats struct {
	ItemsProcessed  int64      // Total items processed
	ItemsSuccessful int64      // Successfully processed items
	ItemsFailed     int64      // Failed items
	LastProcessedAt time.Time  // When last item was processed
	StartTime       time.Time  // When the worker started
	StopTime        time.Time  // When the worker stopped
	ProcessingTime  int64      // Total time spent processing in nanoseconds
	IsRunning       bool       // Is the worker currently running
	mu              sync.Mutex // Mutex for updating stats
}

// NewStats creates a new Stats instance
func NewStats() *Stats {
	return &Stats{
		ItemsProcessed:  0,
		ItemsSuccessful: 0,
		ItemsFailed:     0,
		ProcessingTime:  0,
		IsRunning:       false,
	}
}

// RecordStart marks the worker as started
func (s *Stats) RecordStart() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.StartTime = time.Now()
	s.IsRunning = true
}

// RecordStop marks the worker as stopped
func (s *Stats) RecordStop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.StopTime = time.Now()
	s.IsRunning = false
}

// RecordItemProcessed records a successful item processing
func (s *Stats) RecordItemProcessed() {
	atomic.AddInt64(&s.ItemsProcessed, 1)
	atomic.AddInt64(&s.ItemsSuccessful, 1)
	atomic.AddInt64(&s.ProcessingTime, int64(time.Since(s.StartTime)))

	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastProcessedAt = time.Now()
}

// RecordItemFailed records a failed item processing
func (s *Stats) RecordItemFailed() {
	atomic.AddInt64(&s.ItemsProcessed, 1)
	atomic.AddInt64(&s.ItemsFailed, 1)
	atomic.AddInt64(&s.ProcessingTime, int64(time.Since(s.StartTime)))

	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastProcessedAt = time.Now()
}

// GetSnapshot returns a copy of the current stats
func (s *Stats) GetSnapshot() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	itemsProcessed := atomic.LoadInt64(&s.ItemsProcessed)
	itemsSuccessful := atomic.LoadInt64(&s.ItemsSuccessful)
	itemsFailed := atomic.LoadInt64(&s.ItemsFailed)
	processingTime := atomic.LoadInt64(&s.ProcessingTime)

	return Stats{
		ItemsProcessed:  itemsProcessed,
		ItemsSuccessful: itemsSuccessful,
		ItemsFailed:     itemsFailed,
		LastProcessedAt: s.LastProcessedAt,
		ProcessingTime:  processingTime,
		StartTime:       s.StartTime,
		StopTime:        s.StopTime,
		IsRunning:       s.IsRunning,
	}
}

// GetTotalRuntime returns the total runtime of the worker
func (s *Stats) GetTotalRuntime() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsRunning {
		return time.Since(s.StartTime)
	}

	if s.StopTime.IsZero() {
		return 0
	}

	return s.StopTime.Sub(s.StartTime)
}
