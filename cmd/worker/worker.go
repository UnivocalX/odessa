package main

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/UnivocalX/odessa/internal/tasks"
)

type WorkerConfig struct {
	PollInterval    time.Duration
	MaxPollInterval time.Duration
	Concurrency     int
	MaxAttempts     int
	DrainTimeout    time.Duration
}

type Worker struct {
	cfg      WorkerConfig
	handlers []tasks.Handler
}

func NewWorker(cfg WorkerConfig) *Worker {
	return &Worker{
		cfg: cfg,
	}
}

// Register adds a task handler.
func (w *Worker) Register(handler tasks.Handler) {
	w.handlers = append(w.handlers, handler)
}

func (w *Worker) Run(ctx context.Context) {
	interval := w.cfg.PollInterval
	consecutiveEmpty := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}

		claimed := w.poll(ctx)

		if claimed > 0 {
			consecutiveEmpty = 0
			interval = w.cfg.PollInterval
			slog.Info("poll cycle", "claimed", claimed)
		} else {
			consecutiveEmpty++
			interval = w.backoff(consecutiveEmpty)
			slog.Debug("poll cycle idle", "backoff", interval)
		}
	}
}

// backoff returns an exponential backoff duration with jitter.
func (w *Worker) backoff(consecutive int) time.Duration {
	exp := math.Pow(2, float64(consecutive))
	base := time.Duration(exp) * w.cfg.PollInterval
	if base > w.cfg.MaxPollInterval {
		base = w.cfg.MaxPollInterval
	}
	// Add up to 25% jitter.
	jitter := time.Duration(rand.Int64N(int64(base) / 4))
	return base + jitter
}

// poll claims and processes jobs from all handlers. Returns total jobs claimed.
func (w *Worker) poll(ctx context.Context) int {
	total := 0
	var wg sync.WaitGroup

	for _, h := range w.handlers {
		jobs, err := h.Claim(ctx, w.cfg.Concurrency)
		if err != nil {
			slog.Error("claim jobs", "type", h.Type(), "error", err)
			continue
		}

		total += len(jobs)

		for _, job := range jobs {
			select {
			case <-ctx.Done():
				wg.Wait()
				return total
			default:
			}

			wg.Add(1)
			go func(j tasks.Job, handler tasks.Handler) {
				defer wg.Done()
				w.process(ctx, j, handler)
			}(job, h)
		}
	}

	wg.Wait()
	return total
}

func (w *Worker) process(ctx context.Context, job tasks.Job, handler tasks.Handler) {
	results, err := handler.Handle(ctx, job)
	if err != nil {
		// Cancelled scans already have the correct status in the DB; skip complete/fail.
		if errors.Is(err, tasks.ErrScanCancelled) {
			slog.Info("job cancelled", "type", job.Type, "job_id", job.ID)
			return
		}
		slog.Error("job failed", "type", job.Type, "job_id", job.ID, "attempt", job.Attempt+1, "error", err)
		if failErr := handler.Fail(ctx, job.ID, w.cfg.MaxAttempts); failErr != nil {
			slog.Error("mark job failed", "type", job.Type, "job_id", job.ID, "error", failErr)
		}
		return
	}

	if err := handler.Complete(ctx, job.ID, results); err != nil {
		slog.Error("mark job complete", "type", job.Type, "job_id", job.ID, "error", err)
	}
}
