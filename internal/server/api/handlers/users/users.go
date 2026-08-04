package users

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func HandleListUsers(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "list users start")

		users, err := svc.ListUsers(r.Context())
		if err != nil {
			utils.HandleError(w, r, err)
			return
		}

		resp := dto.ListUsersResponse{Users: make([]dto.UserResponse, len(users))}
		for i, user := range users {
			resp.Users[i] = dto.UserResponse{
				ID:         user.ID,
				Name:       user.Name,
				Email:      user.Email,
				Role:       user.Role,
				DisabledAt: user.DisabledAt,
				CreatedAt:  user.CreatedAt,
			}
		}

		slog.InfoContext(r.Context(), "list users success", "count", len(resp.Users))
		utils.RespondOK(w, r, resp)
	}
}

func HandleCreateUser(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CreateUserRequest
		if err := utils.Decode(r, &req); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "create user attempt", "email", req.Email)

		if _, err := svc.CreateUser(r.Context(), req.Name, req.Email, req.Password); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "create user success", "email", req.Email)
		utils.Respond(w, r, http.StatusCreated, dto.Response{Message: "user created"})
	}
}

func HandleDeleteUser(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
		if err != nil || id == 0 {
			slog.WarnContext(r.Context(), "invalid user id for delete", "raw", r.PathValue("id"))
			utils.RespondBadRequest(w, r, errors.New("invalid user id"))
			return
		}

		slog.InfoContext(r.Context(), "delete user attempt", "user_id", id)

		if err := svc.DeleteUser(r.Context(), uint(id)); err != nil {
			utils.HandleError(w, r, err)
			return
		}

		slog.InfoContext(r.Context(), "delete user success", "user_id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}
