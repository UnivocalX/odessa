package v1

import (
	"log/slog"
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

		slog.InfoContext(r.Context(), "register origin", "uri", req.URI)

		_, err := svc.RegisterOrigin(r.Context(), req.URI, req.Rules)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "register origin success", "uri", req.URI)
		utils.RespondOK(w, r, dto.Response{Message: "successful registered origin"})
	}
}
