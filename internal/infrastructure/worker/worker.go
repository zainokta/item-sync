package worker

import (
	"context"
	"sync"
	"time"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/pkg/logger"
)

type Job interface {
	Name() string
	Execute(ctx context.Context) error
}

type Scheduler struct {
	config   config.WorkerConfig
	logger   logger.Logger
	jobs     map[string]Job
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	running  bool
}

func NewScheduler(config config.WorkerConfig, logger logger.Logger) *Scheduler {
	return &Scheduler{
		config:   config,
		logger:   logger,
		jobs:     make(map[string]Job),
		stopChan: make(chan struct{}),
	}
}

func (s *Scheduler) RegisterJob(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.Name()] = job
	s.logger.Info("Job registered", "name", job.Name())
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	if !s.config.Enabled {
		s.logger.Info("Worker scheduler is disabled")
		return nil
	}

	s.logger.Info("Starting worker scheduler", "sync_interval", s.config.SyncInterval)

	ticker := time.NewTicker(s.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Worker scheduler stopping due to context cancellation")
			return ctx.Err()
		case <-s.stopChan:
			s.logger.Info("Worker scheduler stopping")
			return nil
		case <-ticker.C:
			s.executeJobs(ctx)
		}
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	s.logger.Info("Worker scheduler stopped")
}

func (s *Scheduler) executeJobs(ctx context.Context) {
	s.mu.RLock()
	jobs := make([]Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	s.mu.RUnlock()

	if len(jobs) == 0 {
		s.logger.Debug("No jobs registered, skipping execution")
		return
	}

	s.logger.Info("Starting job execution", "job_count", len(jobs))

	for _, job := range jobs {
		s.wg.Add(1)
		go func(j Job) {
			defer s.wg.Done()
			s.executeJob(ctx, j)
		}(job)
	}
}

func (s *Scheduler) executeJob(ctx context.Context, job Job) {
	jobCtx, cancel := context.WithTimeout(ctx, s.config.JobTimeout)
	defer cancel()

	startTime := time.Now()
	s.logger.Info("Executing job", "name", job.Name())

	err := job.Execute(jobCtx)
	executionTime := time.Since(startTime)

	if err != nil {
		s.logger.Error("Job execution failed",
			"name", job.Name(),
			"error", err,
			"execution_time", executionTime)
	} else {
		s.logger.Info("Job execution completed successfully",
			"name", job.Name(),
			"execution_time", executionTime)
	}
}
