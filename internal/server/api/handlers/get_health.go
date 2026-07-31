package handlers

import (
	"net/http"

	"example.com/aether/internal/server/api/utils"
	"example.com/aether/pkg/dto"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	utils.RespondOK(w, r, dto.Response{Message: "service is healthy"})
}
