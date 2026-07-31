package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleGetOriginScan(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		if rawID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing task id", service.ErrValidation))
			return
		}

		id, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid task id", service.ErrValidation))
			return

		}

		task, err := svc.RetriveTask(r.Context(), uint(id))
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
