package api

import (
	"net/http"
	"strings"

	"example.com/aether/internal/server/api/handlers"
	v1handlers "example.com/aether/internal/server/api/handlers/v1"
	"example.com/aether/internal/server/api/middleware"
	"example.com/aether/internal/server/api/utils"
	"example.com/aether/internal/service"
)

const APIPrefix = "api"

func Route(method string, parts ...string) string {
	parts = append([]string{APIPrefix}, parts...)

	path := strings.Trim(
		strings.TrimSpace(strings.Join(parts, "/")),
		"/\\",
	)

	return method + " /" + path
}

func V1Route(method string, parts ...string) string {
	return Route(method, append([]string{"v1"}, parts...)...)
}

func Register(mux *http.ServeMux, svc *service.Service) http.Handler {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondNotFound(w, r)
	})

	mux.HandleFunc(Route(http.MethodGet, "health"), handlers.HandleHealth)
	mux.HandleFunc(V1Route(http.MethodPost, "origins"), v1handlers.HandlePostOrigin(svc))
	mux.HandleFunc(V1Route(http.MethodGet, "origins"), v1handlers.HandleListOrigins(svc))
	mux.HandleFunc(V1Route(http.MethodPost, "origins/{id}/scan"), v1handlers.HandlePostOriginScan(svc))
	mux.HandleFunc(V1Route(http.MethodGet, "origins/{id}/scan"), v1handlers.HandleGetOriginScan(svc))

	return middleware.Register(mux)
}
