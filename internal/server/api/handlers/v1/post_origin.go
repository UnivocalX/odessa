package v1

import (
	"net/http"

	"example.com/aether/internal/server/api/utils"
	"example.com/aether/internal/service"
	"example.com/aether/pkg/dto"
)

func HandlePostOrigin(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PostOriginRequest

		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		_, err := svc.RegisterOrigin(r.Context(), req.URI)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		utils.RespondOK(w, r, dto.Response{Message: "successful registered origin"})
	}
}
