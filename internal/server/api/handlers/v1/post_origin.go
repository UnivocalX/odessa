package v1

import (
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandlePostOrigin(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PostOriginRequest

		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		_, err := svc.RegisterOrigin(r.Context(), req.URI, req.Rules)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		utils.RespondOK(w, r, dto.Response{Message: "successful registered origin"})
	}
}
