package v1

import (
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleListLabels(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
				ID:        l.ID,
				Name: l.Name,
				Description: l.Description,
			}
		}

		utils.RespondOK(w, r, resp)
	}
}
