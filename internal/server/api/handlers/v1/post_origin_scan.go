package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"example.com/aether/internal/server/api/utils"
	"example.com/aether/internal/service"
	"example.com/aether/pkg/dto"
)

func HandlePostOriginScan(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		if rawID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing task id", service.ErrValidation))
			return
		}

		originID, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid task id", service.ErrValidation))
			return
		}

		task, err := svc.CreateScanOriginTask(r.Context(), uint(originID))
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		utils.RespondOK(w, r, dto.TaskResponse{
			ID:        task.ID,
			Type:      string(task.Type),
			Status:    string(task.Status),
			CreatedAt: task.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
}
