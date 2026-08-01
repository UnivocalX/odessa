package utils

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/pkg/dto"
)

func Respond(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.ErrorContext(r.Context(), "utils: respond: encode", "error", err)
	}
}

func RespondOK(w http.ResponseWriter, r *http.Request, body any) {
	Respond(w, r, http.StatusOK, body)
}

func RespondNotFound(w http.ResponseWriter, r *http.Request) {
	Respond(w, r, http.StatusNotFound, dto.Response{Message: "not found"})
}

func RespondBadRequest(w http.ResponseWriter, r *http.Request, e error) {
	Respond(w, r, http.StatusBadRequest, dto.ErrorResponse{
		Response: dto.Response{Message: "invalid request"},
		Error:    e.Error(),
	})
}

func RespondInternalError(w http.ResponseWriter, r *http.Request) {
	Respond(w, r, http.StatusInternalServerError, dto.Response{Message: "internal server error"})
}

func RespondUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	Respond(w, r, http.StatusUnauthorized, dto.ErrorResponse{
		Response: dto.Response{Message: "unauthorized"},
		Error:    message,
	})
}

func RespondForbidden(w http.ResponseWriter, r *http.Request, message string) {
	Respond(w, r, http.StatusForbidden, dto.ErrorResponse{
		Response: dto.Response{Message: "forbidden"},
		Error:    message,
	})
}

func RespondConflict(w http.ResponseWriter, r *http.Request, e error) {
	Respond(w, r, http.StatusConflict, dto.ErrorResponse{
		Response: dto.Response{Message: "conflict"},
		Error:    e.Error(),
	})
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		Respond(w, r, http.StatusNotFound, dto.ErrorResponse{
			Response: dto.Response{Message: "not found"},
			Error:    err.Error(),
		})
	case errors.Is(err, service.ErrValidation):
		RespondBadRequest(w, r, err)
	case errors.Is(err, service.ErrAlreadyExists):
		RespondConflict(w, r, err)
	case errors.Is(err, service.ErrInvalidCredentials):
		Respond(w, r, http.StatusUnauthorized, dto.ErrorResponse{
			Response: dto.Response{Message: "unauthorized"},
			Error:    err.Error(),
		})
	default:
		RespondInternalError(w, r)
	}
}
