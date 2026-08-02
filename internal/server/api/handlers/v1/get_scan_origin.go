package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleGetScanOrigin(svc *service.BlobService) http.HandlerFunc {
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

		scan, err := svc.RetrieveScanOrigin(r.Context(), uint(oid))
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		utils.RespondOK(w, r, dto.ScanOriginResponse{
			ID:        scan.ID,
			Status:    string(scan.Status),
			CreatedAt: scan.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
}
