// Package cron provides a cron-expression job scheduler for Nexgou applications.
//
// It wraps github.com/robfig/cron/v3 and exposes named jobs with add/remove/list support.
//
// Usage:
//
//	func (s *ReportService) OnStart() {
//	    s.cron.AddJob("daily-report", "0 0 * * *", func() {
//	        s.generateDailyReport()
//	    })
//	    s.cron.Start()
//	}
package cron

import (
	"errors"
	"sync"

	cronlib "github.com/robfig/cron/v3"
	"github.com/nexgou/server/src/logger"
)

// Job represents a registered cron job.
type Job struct {
	Name       string
	Expression string
	EntryID    cronlib.EntryID
}

// CronService manages named cron jobs using standard cron expressions.
//
// Cron expression format (6 fields including seconds):
//
//	┌───────────── second (0–59)
//	│ ┌───────────── minute (0–59)
//	│ │ ┌───────────── hour (0–23)
//	│ │ │ ┌───────────── day of month (1–31)
//	│ │ │ │ ┌───────────── month (1–12)
//	│ │ │ │ │ ┌───────────── day of week (0–6, Sunday=0)
//	│ │ │ │ │ │
//	* * * * * *
//
// Use "@every 5m", "@hourly", "@daily", "@weekly", "@monthly" as shorthand.
type CronService struct {
	c    *cronlib.Cron
	mu   sync.RWMutex
	jobs map[string]*Job
	log  *logger.ScopedLogger
}

// NewCronService creates a new CronService with second-level precision.
// Depends on *logger.LoggerService.
func NewCronService(log *logger.LoggerService) *CronService {
	return &CronService{
		c:    cronlib.New(cronlib.WithSeconds()),
		jobs: make(map[string]*Job),
		log:  log.WithContext("CronService"),
	}
}

// AddJob registers a new named cron job with the given expression.
// Returns an error if the name is already taken or the expression is invalid.
func (s *CronService) AddJob(name, expr string, fn func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.jobs[name]; exists {
		return errors.New("nexgou/cron: job already registered: " + name)
	}
	id, err := s.c.AddFunc(expr, fn)
	if err != nil {
		return err
	}
	s.jobs[name] = &Job{Name: name, Expression: expr, EntryID: id}
	s.log.Info("job registered", "name", name, "expr", expr)
	return nil
}

// RemoveJob removes a named job. If the job is not found, this is a no-op.
func (s *CronService) RemoveJob(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[name]
	if !ok {
		return
	}
	s.c.Remove(job.EntryID)
	delete(s.jobs, name)
	s.log.Info("job removed", "name", name)
}

// Jobs returns a snapshot of all registered jobs.
func (s *CronService) Jobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, j)
	}
	return out
}

// Start begins executing scheduled jobs. Safe to call multiple times.
func (s *CronService) Start() {
	s.c.Start()
	s.log.Info("scheduler started")
}

// Stop halts the scheduler. Running jobs are allowed to complete.
func (s *CronService) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
	s.log.Info("scheduler stopped")
}
