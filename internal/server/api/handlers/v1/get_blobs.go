package v1

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

// HandleGetBlob returns a handler that retrieves a single blob by hash.
func HandleGetBlob(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue("hash")
		if hash == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing blob hash", service.ErrValidation))
			return
		}

		slog.InfoContext(r.Context(), "get blob", "hash", hash)

		b, err := svc.RetrieveBlobByHash(r.Context(), hash)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		resp := buildBlobResponse(*b)
		slog.InfoContext(r.Context(), "get blob success", "hash", resp.Hash, "id", resp.ID)
		utils.RespondOK(w, r, resp)
	}
}

// HandleListBlobs returns a handler that lists blobs with cursor-based
// pagination using query parameters.
func HandleListBlobs(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		var opts []repository.SearchOption

		if cursor, err := parseUintParam(query, "cursor"); err != nil {
			utils.RespondBadRequest(w, r, err)
			return
		} else if cursor > 0 {
			opts = append(opts, repository.WithCursor(cursor))
		}

		if limit, err := parseIntParam(query, "limit"); err != nil {
			utils.RespondBadRequest(w, r, err)
			return
		} else if limit > 0 {
			opts = append(opts, repository.WithLimit(limit))
		}

		slog.InfoContext(r.Context(), "list blobs", "cursor", query.Get("cursor"), "limit", query.Get("limit"))

		result, err := svc.SearchBlobs(r.Context(), opts...)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		resp := buildSearchBlobsResponse(result.Blobs, result.NextCursor, result.HasMore)
		slog.InfoContext(r.Context(), "list blobs success", "count", len(resp.Blobs), "next_cursor", resp.NextCursor, "has_more", resp.HasMore)
		utils.RespondOK(w, r, resp)
	}
}

func HandleSearchBlobs(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SearchBlobsRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		var opts []repository.SearchOption

		// Include filters
		if len(req.Include.Hashes) > 0 {
			opts = append(opts, repository.WithHashes(req.Include.Hashes...))
		}
		if len(req.Include.MimeTypes) > 0 {
			opts = append(opts, repository.WithMimeTypes(req.Include.MimeTypes...))
		}
		if len(req.Include.Labels) > 0 {
			opts = append(opts, repository.WithLabels(req.Include.Labels...))
		}
		if len(req.Include.LabelValues) > 0 {
			opts = append(opts, repository.WithLabelValues(req.Include.LabelValues))
		}
		if req.Include.URIPattern != "" {
			opts = append(opts, repository.WithURIPattern(req.Include.URIPattern))
		}

		// Exclude filters
		if len(req.Exclude.Hashes) > 0 {
			opts = append(opts, repository.WithoutHashes(req.Exclude.Hashes...))
		}
		if len(req.Exclude.MimeTypes) > 0 {
			opts = append(opts, repository.WithoutMimeTypes(req.Exclude.MimeTypes...))
		}
		if len(req.Exclude.Labels) > 0 {
			opts = append(opts, repository.WithoutLabels(req.Exclude.Labels...))
		}
		if len(req.Exclude.LabelValues) > 0 {
			opts = append(opts, repository.WithoutLabelValues(req.Exclude.LabelValues))
		}
		if req.Exclude.URIPattern != "" {
			opts = append(opts, repository.WithoutURIPattern(req.Exclude.URIPattern))
		}

		if req.MinSize != nil {
			opts = append(opts, repository.WithMinSize(*req.MinSize))
		}
		if req.MaxSize != nil {
			opts = append(opts, repository.WithMaxSize(*req.MaxSize))
		}
		if req.Cursor > 0 {
			opts = append(opts, repository.WithCursor(req.Cursor))
		}
		if req.Limit > 0 {
			opts = append(opts, repository.WithLimit(req.Limit))
		}

		result, err := svc.SearchBlobs(r.Context(), opts...)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		resp := buildSearchBlobsResponse(result.Blobs, result.NextCursor, result.HasMore)
		utils.RespondOK(w, r, resp)
	}
}

func parseUintParam(values url.Values, name string) (uint, error) {
	value := values.Get(name)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("invalid %s query parameter: %w", name, err)
	}
	return uint(parsed), nil
}

func parseIntParam(values url.Values, name string) (int, error) {
	value := values.Get(name)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s query parameter: %w", name, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}

func buildSearchBlobsResponse(blobs []repository.Blob, nextCursor uint, hasMore bool) dto.SearchBlobsResponse {
	resp := dto.SearchBlobsResponse{
		Blobs:      make([]dto.BlobResponse, len(blobs)),
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	for i, b := range blobs {
		resp.Blobs[i] = buildBlobResponse(b)
	}

	return resp
}

func buildBlobResponse(b repository.Blob) dto.BlobResponse {
	resp := dto.BlobResponse{
		ID:       b.ID,
		Hash:     b.Hash,
		MimeType: b.MimeType,
		Size:     b.Size,
	}
	// storage locations
	for _, loc := range b.Locations {
		resp.Locations = append(resp.Locations, string(loc.URI))
	}
	// labels
	for _, bl := range b.Labels {
		resp.Labels = append(resp.Labels, dto.BlobLabelResponse{
			Name:  bl.Label.Name,
			Value: bl.Value,
		})
	}
	return resp
}