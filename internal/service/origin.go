package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/go-playground/validator/v10"
)

func (s *BlobService) ListOrigins(ctx context.Context) ([]repository.Origin, error) {
	origins, err := s.repo.ListOrigins(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list origins", "error", err)
		return nil, err
	}
	return origins, nil
}

// RegisterOrigin validates the URI resolves to a configured storage backend,
// checks that the location is accessible, then persists the origin.
func (s *BlobService) RegisterOrigin(ctx context.Context, uri string) (*repository.Origin, error) {
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

func (s *BlobService) RetrieveScanOrigin(ctx context.Context, oid uint) (*repository.ScanOrigin, error) {
	scan, err := s.repo.GetScanOrigin(ctx, oid)
	if err != nil {
		return nil, fmt.Errorf("%w: scan origin %d", ErrNotFound, oid)
	}
	return scan, nil
}

func (s *BlobService) NewScanOrigin(ctx context.Context, oid uint) (*repository.ScanOrigin, error) {
	// Verify the origin exists.
	_, err := s.repo.GetOrigin(ctx, oid)
	if err != nil {
		return nil, fmt.Errorf("%w: origin %d", ErrNotFound, oid)
	}

	task, err := s.repo.CreateScanOrigin(ctx, oid)

	if err != nil {
		slog.ErrorContext(ctx, "failed to create scan origin", "origin_id", oid, "error", err)
		return nil, err
	}

	return task, nil
}
