package service

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/core"
	"github.com/UnivocalX/odessa/internal/repository"
)

func (s *BlobService) NewDataset(ctx context.Context, name string, description string) (*repository.Dataset, error) {
	dataset, err := s.repo.CreateDataset(ctx, name, description)

	if err != nil {
		if errors.Is(err, core.ErrAlreadyExists) {
			return nil, fmt.Errorf("%w: dataset %q", core.ErrAlreadyExists, name)
		}
		slog.ErrorContext(ctx, "failed to create dataset", "dataset_name", name, "error", err)
		return nil, err
	}

	return dataset, nil
}

func (s *BlobService) ListDatasets(ctx context.Context) ([]repository.Dataset, error) {
	datasets, err := s.repo.ListDatasets(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list datasets", "error", err)
		return nil, err
	}
	return datasets, nil
}

func (s *BlobService) NewDatasetVersion(ctx context.Context, id uint, commit string, blobIDs []uint) (*repository.DatasetVersion, error) {
	version, err := s.repo.CreateDatasetVersion(ctx, id, commit)

	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return nil, fmt.Errorf("%w: dataset %d", core.ErrNotFound, id)
		}
		if validationErr := mapValidationError(err); validationErr != nil {
			return nil, validationErr
		}

		slog.ErrorContext(ctx, "failed to create dataset version", "dataset_id", id, "error", err)
		return nil, err
	}

	_, err = s.repo.BatchAssociateBlobsToDatasetVersion(ctx, version.ID, blobIDs)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return nil, fmt.Errorf("%w: dataset version %d", core.ErrNotFound, version.ID)
		}
		if validationErr := mapValidationError(err); validationErr != nil {
			return nil, validationErr
		}

		slog.ErrorContext(ctx, "failed to create dataset version", "dataset_id", id, "error", err)
		return nil, err
	}

	return version, nil
}

func (s *BlobService) ListDatasetVersions(ctx context.Context, datasetID uint) ([]repository.DatasetVersion, error) {
	versions, err := s.repo.ListDatasetVersions(ctx, datasetID)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return nil, fmt.Errorf("%w: dataset %d", core.ErrNotFound, datasetID)
		}
		slog.ErrorContext(ctx, "failed to list dataset versions", "dataset_id", datasetID, "error", err)
		return nil, err
	}
	return versions, nil
}

func (s *BlobService) RetrieveDatasetVersion(ctx context.Context, datasetID uint, versionID uint) (*repository.DatasetVersion, error) {
	version, err := s.repo.GetDatasetVersion(ctx, datasetID, versionID)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return nil, fmt.Errorf("%w: dataset %d version %d", core.ErrNotFound, datasetID, versionID)
		}
		slog.ErrorContext(ctx, "failed to retrieve dataset version", "dataset_id", datasetID, "version_id", versionID, "error", err)
		return nil, err
	}

	return version, nil
}

// RetrieveDatasetVersionBlobs returns an iterator over batches of blobs in a
// dataset version, transparently paging through the repository under the hood.
// Each yielded slice is one page (up to pageSize blobs). Iteration stops early
// (without error) if the consumer breaks out of range.
func (s *BlobService) RetrieveDatasetVersionBlobs(ctx context.Context, datasetID, versionID uint) iter.Seq2[[]repository.Blob, error] {
	const pageSize = 100

	return func(yield func([]repository.Blob, error) bool) {
		var cursor uint
		for {
			blobs, hasMore, err := s.repo.GetDatasetVersionBlobs(ctx, datasetID, versionID, cursor, pageSize)
			if err != nil {
				if errors.Is(err, core.ErrNotFound) {
					err = fmt.Errorf("%w: dataset %d version %d", core.ErrNotFound, datasetID, versionID)
				}
				yield(nil, err)
				return
			}

			if len(blobs) == 0 {
				return
			}

			if !yield(blobs, nil) {
				return // consumer stopped ranging
			}

			cursor = blobs[len(blobs)-1].ID

			if !hasMore {
				return
			}
		}
	}
}