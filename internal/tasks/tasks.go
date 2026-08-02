package tasks

import (
	"context"
)

// Job represents a single unit of work claimed from a queue.
type Job struct {
	ID      uint
	Type    string
	Attempt int
	Payload any
}

// Handler owns the full lifecycle of a task type: claiming jobs from the
// queue, processing them, and marking them complete or failed.
type Handler interface {
	// Type returns the task type identifier.
	Type() string

	// Claim atomically claims up to `limit` jobs from the queue.
	Claim(ctx context.Context, limit int) ([]Job, error)

	// Handle processes the job. The context is cancelled on shutdown.
	Handle(ctx context.Context, job Job) (any, error)

	// Complete marks a job as successfully completed with optional results.
	Complete(ctx context.Context, id uint, results any) error

	// Fail marks a job as failed. Implementations handle retry logic
	// (re-enqueue if under maxAttempts, otherwise mark permanently failed).
	Fail(ctx context.Context, id uint, maxAttempts int) error
}
