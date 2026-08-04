package auth

import (
	"log/slog"
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandlePasswordResetRequest(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PasswordResetRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "password reset requested", "email", req.Email)

		if err := svc.RequestPasswordReset(r.Context(), req.Email); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "password reset request accepted", "email", req.Email)
		utils.RespondOK(w, r, dto.Response{Message: "if the account exists, reset instructions will be sent"})
	}
}

func HandlePasswordResetConfirm(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PasswordResetConfirmRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "password reset confirm attempt")

		if err := svc.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "password reset confirm success")
		utils.RespondOK(w, r, dto.Response{Message: "password reset"})
	}
}
