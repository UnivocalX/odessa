package v1

import (
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleListOrigins(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origins, err := svc.ListOrigins(r.Context())
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		resp := dto.ListOriginsResponse{
			Origins: make([]dto.OriginResponse, len(origins)),
		}
		for i, o := range origins {
			resp.Origins[i] = dto.OriginResponse{
				ID:        o.ID,
				URI:       string(o.URI),
				CreatedAt: o.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}
		}

		utils.RespondOK(w, r, resp)
	}
}
