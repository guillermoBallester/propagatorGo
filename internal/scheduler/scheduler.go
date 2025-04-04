package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/guillermoballester/propagatorGo/internal/config"

	"github.com/robfig/cron/v3"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	StatusIdle      JobStatus = "idle"
	StatusRunning   JobStatus = "running"
	StatusSucceeded JobStatus = "succeeded"
	StatusFailed    JobStatus = "failed"
)

// Job represents a schedulable task
type Job struct {
	cronID      cron.EntryID
	Name        string
	Func        func(ctx context.Context) error
	Timeout     time.Duration
	Status      JobStatus
	LastRun     time.Time
	NextRun     time.Time
	LastError   error
	LastRunTime time.Duration
}

// Scheduler manages scheduled jobs using robfig/cron
type Scheduler struct {
	cron      *cron.Cron
	jobs      map[string]*Job
	jobsMutex sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	config    *config.SchedulerConfig
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.SchedulerConfig) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		jobs:   make(map[string]*Job),
		ctx:    ctx,
		cancel: cancel,
		config: cfg,
	}
}

// AddJob schedules a new job with a cron expression
func (s *Scheduler) AddJob(name, cronExpr string, timeout time.Duration, jobFunc func(ctx context.Context) error) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job '%s' already exists", name)
	}

	job := &Job{
		Name:    name,
		Func:    jobFunc,
		Timeout: timeout,
		Status:  StatusIdle,
	}

	parser := cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)

	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	job.NextRun = schedule.Next(time.Now())

	entryID, addErr := s.cron.AddFunc(cronExpr, func() {
		s.executeJob(name)
	})
	if addErr != nil {
		return fmt.Errorf("failed to schedule job: %w", addErr)
	}

	job.cronID = entryID
	s.jobs[name] = job
	return nil
}

// executeJob runs a job and updates its status
func (s *Scheduler) executeJob(name string) {
	s.jobsMutex.Lock()
	job, exists := s.jobs[name]
	if !exists {
		s.jobsMutex.Unlock()
		return
	}

	job.Status = StatusRunning
	job.LastRun = time.Now()
	s.jobsMutex.Unlock()

	ctx := s.ctx
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(s.ctx, job.Timeout)
		defer cancel()
	}

	startTime := time.Now()
	err := job.Func(ctx)
	elapsed := time.Since(startTime)

	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	job, exists = s.jobs[name]
	if !exists {
		return
	}

	job.LastRunTime = elapsed
	if err != nil {
		job.Status = StatusFailed
		job.LastError = err
	} else {
		job.Status = StatusSucceeded
		job.LastError = nil
	}

	entry := s.cron.Entry(job.cronID)
	job.NextRun = entry.Schedule.Next(time.Now())
}

// RunJob executes a job immediately, regardless of its schedule
func (s *Scheduler) RunJob(name string) error {
	s.jobsMutex.RLock()
	job, exists := s.jobs[name]
	s.jobsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("job '%s' not found", name)
	}

	// Only run if job is not already running
	s.jobsMutex.RLock()
	isRunning := job.Status == StatusRunning
	s.jobsMutex.RUnlock()

	if isRunning {
		return fmt.Errorf("job '%s' is already running", name)
	}

	go s.executeJob(name)
	return nil
}

// Start begins the scheduler
func (s *Scheduler) Start() {
	s.cron.Start()
	log.Println("Scheduler started")
}

// Stop gracefully shuts down the scheduler
func (s *Scheduler) Stop() {
	s.cancel()
	s.cron.Stop()
	log.Println("Scheduler stopped")
}
