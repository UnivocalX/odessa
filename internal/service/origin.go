package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/core"
	"github.com/UnivocalX/odessa/internal/repository"
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
func (s *BlobService) RegisterOrigin(ctx context.Context, uri string, rules *repository.LabelRules) (*repository.Origin, error) {
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

	var rulesJSON json.RawMessage
	if rules != nil {
		rulesJSON, err = json.Marshal(rules)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid rules", core.ErrValidation)
		}
	}

	origin, err := s.repo.CreateOrigin(ctx, uri, rulesJSON)
	if err != nil {
		if errors.Is(err, core.ErrAlreadyExists) {
			return nil, fmt.Errorf("%w: %s", core.ErrAlreadyExists, uri)
		}
		if validationErr := mapValidationError(err); validationErr != nil {
			return nil, validationErr
		}

		slog.ErrorContext(ctx, "failed to persist origin", "uri", uri, "error", err)
		return nil, err
	}

	return origin, nil
}

// UpdateOriginRules replaces the label rules on an existing origin.
func (s *BlobService) UpdateOriginRules(ctx context.Context, oid uint, rules *repository.LabelRules) error {
	_, err := s.repo.GetOrigin(ctx, oid)
	if err != nil {
		return fmt.Errorf("%w: origin %d", core.ErrNotFound, oid)
	}

	var rulesJSON json.RawMessage
	if rules != nil {
		rulesJSON, err = json.Marshal(rules)
		if err != nil {
			return fmt.Errorf("%w: invalid rules", core.ErrValidation)
		}
	}

	if err := s.repo.UpdateOriginRules(ctx, oid, rulesJSON); err != nil {
		slog.ErrorContext(ctx, "failed to update origin rules", "origin_id", oid, "error", err)
		return err
	}
	return nil
}

func (s *BlobService) RetrieveScanOrigin(ctx context.Context, oid uint) (*repository.ScanOrigin, error) {
	scan, err := s.repo.GetScanOrigin(ctx, oid)
	if err != nil {
		return nil, fmt.Errorf("%w: scan origin %d", core.ErrNotFound, oid)
	}
	return scan, nil
}

func (s *BlobService) NewScanOrigin(ctx context.Context, oid uint, rules *repository.LabelRules) (*repository.ScanOrigin, error) {
	var rulesJSON json.RawMessage
	var err error
	if rules != nil {
		rulesJSON, err = json.Marshal(rules)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid rules", core.ErrValidation)
		}
	}

	scan, err := s.repo.CreateScanOrigin(ctx, oid, rulesJSON)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return nil, fmt.Errorf("%w: origin %d", core.ErrNotFound, oid)
		}
		if errors.Is(err, repository.ErrScanAlreadyRunning) {
			return nil, fmt.Errorf("%w: an active origin scan exist %d", core.ErrAlreadyExists, oid)
		}
		slog.ErrorContext(ctx, "failed to create scan origin", "origin_id", oid, "error", err)
		return nil, err
	}

	return scan, nil
}

// CancelScanOrigin cancels the latest pending or in-progress scan for the given origin.
func (s *BlobService) CancelScanOrigin(ctx context.Context, originID uint) (*repository.ScanOrigin, error) {
	scan, err := s.repo.GetScanOrigin(ctx, originID)
	if err != nil {
		return nil, fmt.Errorf("%w: scan for origin %d", core.ErrNotFound, originID)
	}

	if scan.Status == repository.StatusCompleted || scan.Status == repository.StatusFailed || scan.Status == repository.StatusCancelled {
		return nil, fmt.Errorf("%w: scan %d is already %s", ErrCannotCancel, scan.ID, scan.Status)
	}

	if err := s.repo.UpdateScanOriginStatus(ctx, scan.ID, repository.StatusCancelled); err != nil {
		slog.ErrorContext(ctx, "failed to cancel scan origin", "scan_id", scan.ID, "error", err)
		return nil, err
	}
	scan.Status = repository.StatusCancelled

	// Emit a notification so workers listening via LISTEN can react immediately.
	if err := s.repo.NotifyScanCancelled(ctx, scan.ID); err != nil {
		slog.WarnContext(ctx, "notify scan cancelled failed", "scan_id", scan.ID, "error", err)
	}

	return scan, nil
}
