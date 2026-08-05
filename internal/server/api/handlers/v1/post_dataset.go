package v1

import (
	"log/slog"
	"net/http"

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
		utils.RespondOK(w, r, dto.DatasetResponse{ID: dataset.ID, Name: dataset.Name, Description: dataset.Description})
	}
}


func HandlePostDatasetVersion(svc *service.BlobService) http.HandlerFunc {
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
		utils.RespondOK(w, r, dto.DatasetResponse{ID: dataset.ID, Name: dataset.Name, Description: dataset.Description})
	}
}