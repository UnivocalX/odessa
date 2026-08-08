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

func HandleListOrigins(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "list origins start")

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
				Rules:     o.Rules,
			}
		}

		slog.InfoContext(r.Context(), "list origins success", "count", len(resp.Origins))
		utils.RespondOK(w, r, resp)
	}
}

func HandleGetScanOrigin(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		if rawID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing task id", core.ErrValidation))
			return
		}

		oid, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid task id", core.ErrValidation))
			return
		}

		slog.InfoContext(r.Context(), "retrieve origin scan", "scan_id", rawID)

		scan, err := svc.RetrieveScanOrigin(r.Context(), uint(oid))
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "retrieve origin scan success", "scan_id", scan.ID, "status", string(scan.Status))
		utils.RespondOK(w, r, dto.ScanOriginResponse{
			ID:        scan.ID,
			OriginID:  scan.OriginID,
			Status:    string(scan.Status),
			Attempts:  scan.Attempts,
			Results:   scan.Results,
			CreatedAt: scan.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
}
