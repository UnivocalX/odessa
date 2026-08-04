package v1

import (
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandlePostLabel(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PostLabelRequest

		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		l, err := svc.NewLabel(r.Context(), req.Name, req.Description)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		utils.RespondOK(w, r, dto.LabelResponse{ID: l.ID, Name: l.Name, Description: l.Description})
	}
}
