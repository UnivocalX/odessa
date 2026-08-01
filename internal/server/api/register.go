package api

import (
	"net/http"
	"strings"

	"github.com/UnivocalX/odessa/internal/server/api/handlers"
	"github.com/UnivocalX/odessa/internal/server/api/handlers/auth"
	userhandlers "github.com/UnivocalX/odessa/internal/server/api/handlers/users"
	v1handlers "github.com/UnivocalX/odessa/internal/server/api/handlers/v1"
	"github.com/UnivocalX/odessa/internal/server/api/middleware"
	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
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

func AuthRoute(method string, parts ...string) string {
	return Route(method, append([]string{"auth"}, parts...)...)
}

func Register(mux *http.ServeMux, svc *service.Service, maxRequestBodyBytes int64) http.Handler {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondNotFound(w, r)
	})
	// public
	mux.HandleFunc(Route(http.MethodGet, "health"), handlers.HandleHealth)

	// authentication
	mux.HandleFunc(AuthRoute(http.MethodPost, "login"), auth.HandleLogin(svc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "refresh"), auth.HandleRefreshSession(svc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "password/reset/request"), auth.HandlePasswordResetRequest(svc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "password/reset/confirm"), auth.HandlePasswordResetConfirm(svc))

	mux.Handle(AuthRoute(http.MethodPost, "logout"), auth.HandleLogout(svc))
	mux.Handle(AuthRoute(http.MethodPost, "password/change"), auth.HandleChangePassword(svc))
	mux.Handle(AuthRoute(http.MethodPost, "account/disable"), auth.HandleDisableAccount(svc))

	// user management is authenticated now; authorization policy can be refined later.
	mux.Handle(V1Route(http.MethodGet, "users"), userhandlers.HandleListUsers(svc))
	mux.Handle(V1Route(http.MethodPost, "users"), userhandlers.HandleCreateUser(svc))
	mux.Handle(V1Route(http.MethodDelete, "users/{id}"), userhandlers.HandleDeleteUser(svc))

	// data management
	mux.Handle(V1Route(http.MethodGet, "origins"), v1handlers.HandleListOrigins(svc))
	mux.Handle(V1Route(http.MethodPost, "origins"), v1handlers.HandlePostOrigin(svc))
	mux.Handle(V1Route(http.MethodPost, "origins/{id}/scan"), v1handlers.HandlePostOriginScan(svc))
	mux.Handle(V1Route(http.MethodGet, "origins/{id}/scan"), v1handlers.HandleGetOriginScan(svc))

	// middleware
	authentication := middleware.AuthenticateRoutes(svc)
	authorization := middleware.AuthorizeRoutes(svc)
	limiter := middleware.RequestBodyLimit(maxRequestBodyBytes)

	handler := authentication(authorization(mux))
	handler = limiter(handler)

	return middleware.Register(handler)
}
