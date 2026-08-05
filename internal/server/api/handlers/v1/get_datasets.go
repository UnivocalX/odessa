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

func HandleListDatasets(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "list datasets start")

		datasets, err := svc.ListDatasets(r.Context())
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		resp := dto.ListDatasetsResponse{
			Datasets: make([]dto.DatasetResponse, len(datasets)),
		}
		for i, dataset := range datasets {
			resp.Datasets[i] = dto.DatasetResponse{
				ID:          dataset.ID,
				Name:        dataset.Name,
				Description: dataset.Description,
				CreatedAt:   dataset.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}
		}

		slog.InfoContext(r.Context(), "list datasets success", "count", len(resp.Datasets))
		utils.RespondOK(w, r, resp)
	}
}

func HandleGetDatasetVersions(svc *service.BlobService) http.HandlerFunc {
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

		slog.InfoContext(r.Context(), "list dataset versions", "dataset_id", datasetID)
		versions, err := svc.ListDatasetVersions(r.Context(), uint(datasetID))
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		resp := dto.ListDatasetVersionsResponse{
			Versions: make([]dto.DatasetVersionResponse, len(versions)),
		}
		for i, version := range versions {
			versionWithBlobs, err := svc.RetrieveDatasetVersion(r.Context(), uint(datasetID), version.ID)
			if err != nil {
				utils.HandleError(w, r, err)
				return
			}

			blobIDs := make([]uint, len(versionWithBlobs.Blobs))
			for j, blob := range versionWithBlobs.Blobs {
				blobIDs[j] = blob.ID
			}

			resp.Versions[i] = dto.DatasetVersionResponse{
				ID:        version.ID,
				DatasetID: version.DatasetID,
				Commit:    version.Commit,
				CreatedAt: version.CreatedAt.Format("2006-01-02T15:04:05Z"),
				BlobIDs:   blobIDs,
			}
		}

		slog.InfoContext(r.Context(), "list dataset versions success", "dataset_id", datasetID, "count", len(resp.Versions))
		utils.RespondOK(w, r, resp)
	}
}

func HandleGetDatasetVersion(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawDatasetID := r.PathValue("id")
		if rawDatasetID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing dataset id", core.ErrValidation))
			return
		}

		rawVersionID := r.PathValue("version_id")
		if rawVersionID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing version id", core.ErrValidation))
			return
		}

		datasetID, err := strconv.ParseUint(rawDatasetID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid dataset id", core.ErrValidation))
			return
		}

		versionID, err := strconv.ParseUint(rawVersionID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid version id", core.ErrValidation))
			return
		}

		slog.InfoContext(r.Context(), "get dataset version", "dataset_id", datasetID, "version_id", versionID)
		version, err := svc.RetrieveDatasetVersion(r.Context(), uint(datasetID), uint(versionID))
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		blobIDs := make([]uint, len(version.Blobs))
		for i, blob := range version.Blobs {
			blobIDs[i] = blob.ID
		}

		slog.InfoContext(r.Context(), "get dataset version success", "dataset_id", datasetID, "version_id", versionID)
		utils.RespondOK(w, r, dto.DatasetVersionResponse{
			ID:        version.ID,
			DatasetID: version.DatasetID,
			Commit:    version.Commit,
			CreatedAt: version.CreatedAt.Format("2006-01-02T15:04:05Z"),
			BlobIDs:   blobIDs,
		})
	}
}
