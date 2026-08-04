package service

import (
	"context"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
)

// SearchBlobs returns a single page of blobs matching the given search options.
func (s *BlobService) SearchBlobs(ctx context.Context, opts ...repository.SearchOption) (*repository.BlobSearchResult, error) {
	result, err := s.repo.SearchBlobsPage(ctx, opts...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to search blobs", "error", err)
		return nil, err
	}
	return result, nil
}
