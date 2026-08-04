package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
)

func (s *BlobService) NewLabel(ctx context.Context, name string, description string) (*repository.Label, error) {
	label, err := s.repo.CreateLabel(ctx, name, description)

	if err != nil {
		slog.ErrorContext(ctx, "failed to create label", "label_name", name, "error", err)
		return nil, err
	}

	return label, nil
}

func (s *BlobService) ListLabels(ctx context.Context) ([]repository.Label, error) {
	labels, err := s.repo.ListLabels(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list labels", "error", err)
		return nil, err
	}
	return labels, nil
}

func (s *BlobService) RetrieveLabel(ctx context.Context, id uint) (*repository.Label, error) {
	label, err := s.repo.GetLabel(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: label %d", ErrNotFound, id)
	}
	return label, nil
}
