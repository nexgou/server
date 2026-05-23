package task

import (
	"github.com/nexgou/server/src/logger"
	"github.com/nexgou/server/src/module/cron"
)

// TaskCron schedules periodic maintenance jobs for the task resource.
type TaskCron struct {
	cron    *cron.CronService
	service *TaskService
	log     *logger.ScopedLogger
}

// NewTaskCron registers cron jobs and starts the scheduler.
func NewTaskCron(c *cron.CronService, svc *TaskService, log *logger.LoggerService) *TaskCron {
	tc := &TaskCron{
		cron:    c,
		service: svc,
		log:     log.WithContext("TaskCron"),
	}

	// Every minute: log a quick board snapshot.
	_ = c.AddJob("task.stats", "0 * * * * *", func() {
		total := svc.Count()
		done := svc.CountDone()
		tc.log.Info("board snapshot", "total", total, "done", done, "pending", total-done)
	})

	// Every day at midnight: purge completed tasks.
	_ = c.AddJob("task.cleanup", "0 0 0 * * *", func() {
		n, err := svc.DeleteCompleted()
		if err != nil {
			tc.log.Error("cleanup failed", "err", err)
			return
		}
		tc.log.Info("completed tasks purged", "count", n)
	})

	c.Start()
	return tc
}
