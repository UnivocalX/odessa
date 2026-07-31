package auth

import (
	"net/http"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleLogin(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.LoginRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}
		tokens, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}
		utils.RespondOK(w, r, dto.LoginResponse{AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken, TokenType: "Bearer", ExpiresIn: tokens.ExpiresIn})
	}
}

func HandleRefreshSession(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.RefreshRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}
		tokens, err := svc.RefreshSession(r.Context(), req.RefreshToken)
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}
		utils.RespondOK(w, r, dto.LoginResponse{AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken, TokenType: "Bearer", ExpiresIn: tokens.ExpiresIn})
	}
}
