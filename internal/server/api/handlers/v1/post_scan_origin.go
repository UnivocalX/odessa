package v1

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandlePostOriginScan(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		if rawID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing task id", service.ErrValidation))
			return
		}

		oid, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid task id", service.ErrValidation))
			return
		}

		var req dto.PostScanOriginRequest
		if r.Body != nil && r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				utils.HandleError(w, r, fmt.Errorf("%w: %s", service.ErrValidation, err))
				return
			}
		}

		rulesCount := 0
		if req.Rules != nil {
			rulesCount = len(*req.Rules)
		}
		slog.InfoContext(r.Context(), "create origin scan", "origin_id", rawID, "rules", rulesCount)

		scan, err := svc.NewScanOrigin(r.Context(), uint(oid), req.Rules)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "create origin scan success", "scan_id", scan.ID, "origin_id", rawID)
		utils.RespondOK(w, r, dto.ScanOriginResponse{
			ID:        scan.ID,
			OriginID:  scan.OriginID,
			Status:    string(scan.Status),
			Attempts:  scan.Attempts,
			CreatedAt: scan.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
}

// HandleCancelOriginScan cancels an existing scan.
func HandleCancelOriginScan(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		if rawID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing task id", service.ErrValidation))
			return
		}

		oid, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid task id", service.ErrValidation))
			return
		}

		scan, err := svc.CancelScanOrigin(r.Context(), uint(oid))
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		utils.RespondOK(w, r, dto.ScanOriginResponse{
			ID:        scan.ID,
			OriginID:  scan.OriginID,
			Status:    string(scan.Status),
			Attempts:  scan.Attempts,
			CreatedAt: scan.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
}
