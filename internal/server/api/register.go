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

func Register(mux *http.ServeMux, authSvc *service.AuthService, blobSvc *service.BlobService, maxRequestBodyBytes int64) http.Handler {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondNotFound(w, r)
	})
	// public
	mux.HandleFunc(Route(http.MethodGet, "health"), handlers.HandleHealth)

	// authentication
	mux.HandleFunc(AuthRoute(http.MethodPost, "login"), auth.HandleLogin(authSvc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "refresh"), auth.HandleRefreshSession(authSvc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "password/reset/request"), auth.HandlePasswordResetRequest(authSvc))
	mux.HandleFunc(AuthRoute(http.MethodPost, "password/reset/confirm"), auth.HandlePasswordResetConfirm(authSvc))

	mux.Handle(AuthRoute(http.MethodPost, "logout"), auth.HandleLogout(authSvc))
	mux.Handle(AuthRoute(http.MethodPost, "password/change"), auth.HandleChangePassword(authSvc))
	mux.Handle(AuthRoute(http.MethodPost, "account/disable"), auth.HandleDisableAccount(authSvc))

	// user management is authenticated now; authorization policy can be refined later.
	mux.Handle(V1Route(http.MethodGet, "users"), userhandlers.HandleListUsers(authSvc))
	mux.Handle(V1Route(http.MethodPost, "users"), userhandlers.HandleCreateUser(authSvc))
	mux.Handle(V1Route(http.MethodDelete, "users/{id}"), userhandlers.HandleDeleteUser(authSvc))

	// data management
	mux.Handle(V1Route(http.MethodGet, "origins"), v1handlers.HandleListOrigins(blobSvc))
	mux.Handle(V1Route(http.MethodPost, "origins"), v1handlers.HandlePostOrigin(blobSvc))
	mux.Handle(V1Route(http.MethodPut, "origins/{id}/rules"), v1handlers.HandlePutOriginRules(blobSvc))
	mux.Handle(V1Route(http.MethodPost, "origins/{id}/scan"), v1handlers.HandlePostOriginScan(blobSvc))
	mux.Handle(V1Route(http.MethodGet, "origins/{id}/scan"), v1handlers.HandleGetScanOrigin(blobSvc))
	mux.Handle(V1Route(http.MethodPost, "origins/{id}/scan/cancel"), v1handlers.HandleCancelOriginScan(blobSvc))

	mux.Handle(V1Route(http.MethodGet, "labels"), v1handlers.HandleListLabels(blobSvc))
	mux.Handle(V1Route(http.MethodPost, "labels"), v1handlers.HandlePostLabel(blobSvc))

	mux.Handle(V1Route(http.MethodGet, "blobs"), v1handlers.HandleListBlobs(blobSvc))
	mux.Handle(V1Route(http.MethodGet, "blobs/{hash}"), v1handlers.HandleGetBlob(blobSvc))
	mux.Handle(V1Route(http.MethodPost, "blobs/search"), v1handlers.HandleSearchBlobs(blobSvc))

	mux.Handle(V1Route(http.MethodGet, "datasets"), v1handlers.HandleListDatasets(blobSvc))
	mux.Handle(V1Route(http.MethodPost, "datasets"), v1handlers.HandlePostDataset(blobSvc))
	mux.Handle(V1Route(http.MethodGet, "datasets/{id}/versions"), v1handlers.HandleGetDatasetVersions(blobSvc))
	mux.Handle(V1Route(http.MethodPost, "datasets/{id}/versions"), v1handlers.HandlePostDatasetVersion(blobSvc))
	mux.Handle(V1Route(http.MethodGet, "datasets/{id}/versions/{version_id}"), v1handlers.HandleGetDatasetVersion(blobSvc))
	mux.Handle(V1Route(http.MethodGet, "datasets/{id}/versions/{version_id}/blobs"), v1handlers.HandleGetDatasetVersionBlobs(blobSvc))

	// middleware
	authentication := middleware.AuthenticateRoutes(authSvc)
	authorization := middleware.AuthorizeRoutes(authSvc)
	limiter := middleware.RequestBodyLimit(maxRequestBodyBytes)

	handler := authentication(authorization(mux))
	handler = limiter(handler)

	return middleware.Register(handler)
}
