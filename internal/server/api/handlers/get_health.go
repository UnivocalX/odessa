package handlers

import (
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	utils.RespondOK(w, r, dto.Response{Message: "service is healthy"})
}
