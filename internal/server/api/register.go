package api

import (
	"net/http"
	"strings"

	"github.com/UnivocalX/odessa/internal/server/api/handlers"
	"github.com/UnivocalX/odessa/internal/server/api/handlers/auth"
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

	// authentication
	mux.HandleFunc(AuthRoute(http.MethodPost, "signup"), auth.HandleSignUp(svc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "login"), auth.HandleLogin(svc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "refresh"), auth.HandleRefreshSession(svc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "password/reset/request"), auth.HandlePasswordResetRequest(svc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "password/reset/confirm"), auth.HandlePasswordResetConfirm(svc))

	authenticated := middleware.Authenticate(svc)
	mux.Handle(AuthRoute(http.MethodPost, "logout"), authenticated(auth.HandleLogout(svc)))
	mux.Handle(AuthRoute(http.MethodPost, "password/change"), authenticated(auth.HandleChangePassword(svc)))
	mux.Handle(AuthRoute(http.MethodPost, "account/disable"), authenticated(auth.HandleDisableAccount(svc)))

	// public
	mux.HandleFunc(Route(http.MethodGet, "health"), handlers.HandleHealth)

	// data management
	mux.Handle(V1Route(http.MethodGet, "origins"), authenticated(v1handlers.HandleListOrigins(svc)))
	mux.Handle(V1Route(http.MethodPost, "origins"), authenticated(v1handlers.HandlePostOrigin(svc)))
	mux.Handle(V1Route(http.MethodPost, "origins/{id}/scan"), authenticated(v1handlers.HandlePostOriginScan(svc)))
	mux.Handle(V1Route(http.MethodGet, "origins/{id}/scan"), authenticated(v1handlers.HandleGetOriginScan(svc)))

	return middleware.Register(middleware.RequestBodyLimit(maxRequestBodyBytes)(mux))
}
