package v1

import (
	"log/slog"
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleListLabels(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "list labels start")

		labels, err := svc.ListLabels(r.Context())
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		resp := dto.ListLabelsResponse{
			Labels: make([]dto.LabelResponse, len(labels)),
		}
		for i, l := range labels {
			resp.Labels[i] = dto.LabelResponse{
				ID:          l.ID,
				Name:        l.Name,
				Description: l.Description,
			}
		}

		slog.InfoContext(r.Context(), "list labels success", "count", len(resp.Labels))
		utils.RespondOK(w, r, resp)
	}
}
