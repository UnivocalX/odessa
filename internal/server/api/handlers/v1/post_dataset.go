package v1

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/UnivocalX/odessa/internal/core"
	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandlePostDataset(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PostDatasetRequest

		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "create dataset", "name", req.Name)
		dataset, err := svc.NewDataset(r.Context(), req.Name, req.Description)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "create dataset success", "name", req.Name)
		utils.RespondOK(w, r, dto.DatasetResponse{
			ID:          dataset.ID,
			Name:        dataset.Name,
			Description: dataset.Description,
			CreatedAt:   dataset.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
}

func HandlePostDatasetVersion(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		if rawID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing dataset id", core.ErrValidation))
			return
		}

		datasetID, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid dataset id", core.ErrValidation))
			return
		}

		var req dto.PostDatasetVersionRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "create dataset version", "dataset_id", datasetID)
		version, err := svc.NewDatasetVersion(r.Context(), uint(datasetID), req.Commit, req.BlobIDs)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "create dataset version success", "dataset_id", datasetID, "version_id", version.ID)
		utils.RespondOK(w, r, dto.DatasetVersionResponse{
			ID:        version.ID,
			DatasetID: version.DatasetID,
			Commit:    version.Commit,
			CreatedAt: version.CreatedAt.Format("2006-01-02T15:04:05Z"),
			BlobIDs:   req.BlobIDs,
		})
	}
}
