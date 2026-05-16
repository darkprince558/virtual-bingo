package automation

import (
	"context"
	"time"
)

type JobKind string

const (
	JobPrepareGameContent JobKind = "prepare_game_content"
	JobLockGameContent    JobKind = "lock_game_content"
)

type DueJob struct {
	GameRunID string
	Kind      JobKind
	DueAt     time.Time
}

type Runner interface {
	PrepareGameContent(context.Context, string) error
	LockGameContent(context.Context, string) error
}

type Scheduler struct {
	runner Runner
	now    func() time.Time
}

func NewScheduler(runner Runner, now func() time.Time) Scheduler {
	if now == nil {
		now = time.Now
	}
	return Scheduler{runner: runner, now: now}
}

func (s Scheduler) Due(jobs []DueJob) []DueJob {
	now := s.now()
	due := make([]DueJob, 0, len(jobs))
	for _, job := range jobs {
		if !job.DueAt.After(now) {
			due = append(due, job)
		}
	}

	return due
}

func (s Scheduler) RunDue(ctx context.Context, jobs []DueJob) error {
	for _, job := range s.Due(jobs) {
		switch job.Kind {
		case JobPrepareGameContent:
			if err := s.runner.PrepareGameContent(ctx, job.GameRunID); err != nil {
				return err
			}
		case JobLockGameContent:
			if err := s.runner.LockGameContent(ctx, job.GameRunID); err != nil {
				return err
			}
		}
	}

	return nil
}
