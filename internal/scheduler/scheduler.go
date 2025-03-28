package scheduler

import (
	"context"
	"fmt"
	"log"
	"propagatorGo/internal/config"
	"sync"
	"time"

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

func (s *Scheduler) Initialize() error {
	for _, jobConfig := range s.config.Jobs {
		if !jobConfig.Enabled {
			continue
		}

		placeholderFunc := func(ctx context.Context) error {
			return fmt.Errorf("job implementation not registered")
		}

		err := s.AddJob(jobConfig.Name, jobConfig.CronExpr, jobConfig.Timeout, placeholderFunc)
		if err != nil {
			return fmt.Errorf("cannot add job %s: %w", jobConfig.Name, err)
		}
	}
	log.Println("scheduler initialized")
	return nil
}

// RegisterJobHandler registers implementation for a pre-configured job
func (s *Scheduler) RegisterJobHandler(name string, handler func(ctx context.Context) error) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job '%s' not found", name)
	}

	job.Func = handler
	return nil
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

// getEntryJobName tries to extract job name from cron entry
// This is a helper function since robfig/cron doesn't store the job name directly
func (s *Scheduler) getEntryJobName(entry cron.Entry) string {
	for name, job := range s.jobs {
		if job.cronID == entry.ID {
			return name
		}
	}
	return ""
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

// GetJobStatus returns the current status of a job
func (s *Scheduler) GetJobStatus(name string) (*JobStatus, error) {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	job, exists := s.jobs[name]
	if !exists {
		return nil, fmt.Errorf("job '%s' not found", name)
	}

	return &job.Status, nil
}

// GetAllJobs returns information about all jobs
func (s *Scheduler) GetAllJobs() map[string]Job {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	result := make(map[string]Job)
	for name, job := range s.jobs {
		// Create a copy to avoid mutex issues
		result[name] = *job
	}

	return result
}

// RemoveJob stops and removes a job from the scheduler
func (s *Scheduler) RemoveJob(name string) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	if _, exists := s.jobs[name]; !exists {
		return fmt.Errorf("job '%s' not found", name)
	}

	// Find and remove the job from cron
	for _, entry := range s.cron.Entries() {
		if s.getEntryJobName(entry) == name {
			s.cron.Remove(entry.ID)
			break
		}
	}

	// Remove from jobs map
	delete(s.jobs, name)
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

// PauseJob temporarily disables a job
func (s *Scheduler) PauseJob(name string) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job '%s' not found", name)
	}

	// Find and remove the job from cron
	var entryID cron.EntryID
	var found bool

	for _, entry := range s.cron.Entries() {
		if s.getEntryJobName(entry) == name {
			entryID = entry.ID
			found = true
			break
		}
	}

	if found {
		s.cron.Remove(entryID)
		job.Status = StatusIdle
	}

	return nil
}

// ResumeJob re-enables a paused job
func (s *Scheduler) ResumeJob(name string, cronExpr string) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job '%s' not found", name)
	}

	// Add to cron scheduler again
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeJob(name)
	})
	if err != nil {
		return fmt.Errorf("failed to resume job: %w", err)
	}

	job.cronID = entryID
	job.Status = StatusIdle

	return nil
}
