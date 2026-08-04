package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

// HandlePutOriginRules replaces the label rules on an existing origin.
func HandlePutOriginRules(svc *service.BlobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		if rawID == "" {
			utils.HandleError(w, r, fmt.Errorf("%w: missing origin id", service.ErrValidation))
			return
		}

		oid, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			utils.HandleError(w, r, fmt.Errorf("%w: invalid origin id", service.ErrValidation))
			return
		}

		var req dto.PutOriginRulesRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		if err := svc.UpdateOriginRules(r.Context(), uint(oid), &req.Rules); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		utils.RespondOK(w, r, dto.Response{Message: "rules updated"})
	}
}
