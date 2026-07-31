package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
)

func (s *Service) RetrieveTask(ctx context.Context, id uint) (*repository.Task, error) {
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: task %d", ErrNotFound, id)
	}
	return task, nil
}

func (s *Service) NewScanOriginTask(ctx context.Context, originID uint) (*repository.Task, error) {
	// Verify the origin exists.
	_, err := s.repo.GetOrigin(ctx, originID)
	if err != nil {
		return nil, fmt.Errorf("%w: origin %d", ErrNotFound, originID)
	}

	// Prevent duplicate active scans for the same origin.
	active, err := s.repo.HasActiveScanForOrigin(ctx, originID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check active scan", "origin_id", originID, "error", err)
		return nil, err
	}
	if active {
		return nil, fmt.Errorf("%w: a scan is already active for origin %d", ErrAlreadyExists, originID)
	}

	task, err := s.repo.CreateTask(
		ctx,
		repository.TaskTypeScanOrigin,
		repository.ScanOriginDetails{OriginID: originID},
	)

	if err != nil {
		slog.ErrorContext(ctx, "failed to create scan task", "origin_id", originID, "error", err)
		return nil, err
	}

	return task, nil
}
