package auth

import (
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/middleware"
	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleLogout(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.RefreshRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}
		userID, ok := middleware.UserID(r)
		if !ok {
			utils.RespondUnauthorized(w, r, "missing authenticated user")
			return
		}
		if err := svc.Logout(r.Context(), userID, req.RefreshToken); err != nil {
			utils.HandleError(w, r, err)
			return
		}
		utils.RespondOK(w, r, dto.Response{Message: "logged out"})
	}
}

func HandleChangePassword(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserID(r)
		if !ok {
			utils.Respond(w, r, http.StatusUnauthorized, dto.ErrorResponse{
				Response: dto.Response{Message: "unauthorized"},
				Error:    "missing authenticated user",
			})
			return
		}
		var req dto.ChangePasswordRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}
		if err := svc.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			utils.HandleError(w, r, err)
			return
		}
		utils.RespondOK(w, r, dto.Response{Message: "password changed"})
	}
}

func HandleDisableAccount(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserID(r)
		if !ok {
			utils.RespondUnauthorized(w, r, "missing authenticated user")
			return
		}
		var req dto.DisableAccountRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}
		if err := svc.DisableAccount(r.Context(), userID, req.Password); err != nil {
			utils.HandleError(w, r, err)
			return
		}
		utils.RespondOK(w, r, dto.Response{Message: "account disabled"})
	}
}
