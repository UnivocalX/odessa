package v1

import (
	"net/http"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

// HandleSearchBlobs returns a handler that searches blobs with filtering
// and cursor-based pagination.
func HandleSearchBlobs(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SearchBlobsRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		var opts []repository.SearchOption

		if len(req.Hashes) > 0 {
			opts = append(opts, repository.WithHashes(req.Hashes...))
		}
		if len(req.MimeTypes) > 0 {
			opts = append(opts, repository.WithMimeTypes(req.MimeTypes...))
		}
		if len(req.Labels) > 0 {
			opts = append(opts, repository.WithLabels(req.Labels...))
		}
		if len(req.LabelValues) > 0 {
			opts = append(opts, repository.WithLabelValues(req.LabelValues))
		}
		if req.URIPattern != "" {
			opts = append(opts, repository.WithURIPattern(req.URIPattern))
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

		resp := dto.SearchBlobsResponse{
			Blobs:      make([]dto.BlobResponse, len(result.Blobs)),
			NextCursor: result.NextCursor,
			HasMore:    result.HasMore,
		}
		for i, b := range result.Blobs {
			blob := dto.BlobResponse{
				ID:       b.ID,
				Hash:     b.Hash,
				MimeType: b.MimeType,
				Size:     b.Size,
			}
			for _, bl := range b.Labels {
				blob.Labels = append(blob.Labels, dto.BlobLabelResponse{
					Name:  bl.Label.Name,
					Value: bl.Value,
				})
			}
			resp.Blobs[i] = blob
		}

		utils.RespondOK(w, r, resp)
	}
}
