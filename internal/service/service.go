package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/storage"
	"github.com/go-playground/validator/v10"
)

type Service struct {
	repo *repository.Repository
	reg  *storage.Registry
}

func New(repo *repository.Repository, reg *storage.Registry) *Service {
	return &Service{repo: repo, reg: reg}
}

func (s *Service) ListOrigins(ctx context.Context) ([]repository.Origin, error) {
	origins, err := s.repo.ListOrigins(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list origins", "error", err)
		return nil, err
	}
	return origins, nil
}

// RegisterOrigin validates the URI resolves to a configured storage backend,
// checks that the location is accessible, then persists the origin.
func (s *Service) RegisterOrigin(ctx context.Context, uri string) (*repository.Origin, error) {
	store, err := s.reg.Resolve(uri)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedBackend, uri)
	}

	available, err := store.Available(ctx, uri)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check location availability", "uri", uri, "error", err)
		return nil, fmt.Errorf("%w: %s: %w", ErrLocationNotAccessible, uri, err)
	}
	if !available {
		return nil, fmt.Errorf("%w: %s", ErrLocationNotFound, uri)
	}

	origin, err := s.repo.CreateOrigin(ctx, uri)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, fmt.Errorf("%w: %s", ErrAlreadyExists, uri)
		}

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return nil, fmt.Errorf("%w: %w", ErrValidation, err)
		}

		slog.ErrorContext(ctx, "failed to persist origin", "uri", uri, "error", err)
		return nil, err
	}

	return origin, nil
}

func (s *Service) CreateScanOriginTask(ctx context.Context, originID uint) (*repository.Task, error) {
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

func (s *Service) RetriveTask(ctx context.Context, id uint) (*repository.Task, error) {
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: task %d", ErrNotFound, id)
	}
	return task, nil
}
