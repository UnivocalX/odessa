package service

import (
	"context"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
)

// SearchBlobs returns a single page of blobs matching the given options.
func (s *BlobService) SearchBlobs(ctx context.Context, opts ...repository.SearchOption) (*repository.BlobSearchResult, error) {
	result, err := s.repo.SearchBlobs(ctx, opts...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to search blobs", "error", err)
		return nil, err
	}
	return result, nil
}

// RetrieveBlobByHash returns a blob by its hash.
func (s *BlobService) RetrieveBlobByHash(ctx context.Context, hash string) (*repository.Blob, error) {
	blob, err := s.repo.GetBlobByHash(ctx, hash)
	if err != nil {
		slog.ErrorContext(ctx, "failed to retrieve blob by hash", "error", err, "hash", hash)
		return nil, err
	}
	return blob, nil
}

// RetrieveBlob returns a blob by its ID.
func (s *BlobService) RetrieveBlob(ctx context.Context, id uint) (*repository.Blob, error) {
	blob, err := s.repo.GetBlobByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to retrieve blob", "error", err)
		return nil, err
	}
	return blob, nil
}
