package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/go-playground/validator/v10"
)

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
