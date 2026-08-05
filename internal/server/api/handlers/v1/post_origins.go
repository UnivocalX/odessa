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

func HandlePostOriginScan(svc *service.BlobService) http.HandlerFunc {
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

		var req dto.PostScanOriginRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
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
			utils.HandleError(w, r, fmt.Errorf("%w: missing task id", core.ErrValidation))
			return
		}

		oid, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid task id", core.ErrValidation))
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
