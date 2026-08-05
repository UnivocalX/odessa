package service

import (
	"context"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
)

func (s *BlobService) NewDataset(ctx context.Context, name string, description string) (*repository.Dataset, error) {
	dataset, err := s.repo.CreateDataset(ctx, name, description)

	if err != nil {
		slog.ErrorContext(ctx, "failed to create dataset", "dataset_name", name, "error", err)
		return nil, err
	}

	return dataset, nil
}

func (s *BlobService) NewDatasetVersion(ctx context.Context, id uint, commit string) (*repository.DatasetVersion, error) {
	version, err := s.repo.CreateDatasetVersion(ctx, id, commit)

	if err != nil {
		slog.ErrorContext(ctx, "failed to create dataset version", "dataset_id", id, "error", err)
		return nil, err
	}

	return version, nil
}