package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/UnivocalX/odessa/internal/server/api/utils"
	"github.com/UnivocalX/odessa/internal/service"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func Register(handler http.Handler) http.Handler {
	handler = Logging(handler)
	handler = SecurityHeaders(handler)
	handler = Recovery(handler)
	return handler
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// RequestBodyLimit applies a per-request body limit. Rate limiting and abuse
// protection are intentionally delegated to the API gateway/reverse proxy.
func RequestBodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder wraps http.ResponseWriter to capture the status code,
// since ResponseWriter doesn't expose it after WriteHeader is called.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rec *statusRecorder) WriteHeader(status int) {
	if rec.wroteHeader {
		return
	}
	rec.wroteHeader = true
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *statusRecorder) Write(body []byte) (int, error) {
	if !rec.wroteHeader {
		rec.WriteHeader(http.StatusOK)
	}
	return rec.ResponseWriter.Write(body)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, r)

		slog.Info(
			"http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "path", r.URL.Path)
				utils.RespondInternalError(w, r)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Authenticate validates a Bearer JWT and adds its subject to the request context.
func Authenticate(svc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const unauthorizedMessage = "missing or invalid authentication token"
			header := r.Header.Get("Authorization")
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				respondUnauthorized(w, r, unauthorizedMessage)
				return
			}

			userID, err := svc.ValidateAccessToken(r.Context(), parts[1])
			if err != nil {
				respondUnauthorized(w, r, unauthorizedMessage)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, uint(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthenticateRoutes protects every non-public API route and adds the user ID
// to the request context. Public routes are explicitly listed below.
func AuthenticateRoutes(svc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !requiresAuthentication(r.Method, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			const unauthorizedMessage = "missing or invalid authentication token"
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				respondUnauthorized(w, r, unauthorizedMessage)
				return
			}

			userID, err := svc.ValidateAccessToken(r.Context(), parts[1])
			if err != nil {
				respondUnauthorized(w, r, unauthorizedMessage)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission allows an authenticated user with the requested permission.
func RequirePermission(svc *service.AuthService, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserID(r)
			if !ok || !svc.HasPermission(r.Context(), userID, permission) {
				utils.RespondForbidden(w, r, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthorizeRoutes applies route permissions centrally. Routes without an
// entry are available to any authenticated user.
func AuthorizeRoutes(svc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permission := requiredPermission(r.Method, r.URL.Path)
			if permission == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := UserID(r)
			if !ok || !svc.HasPermission(r.Context(), userID, permission) {
				utils.RespondForbidden(w, r, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func requiresAuthentication(method, path string) bool {
	if strings.HasPrefix(path, "/api/v1/") {
		return true
	}
	return method == http.MethodPost && (path == "/api/auth/logout" ||
		path == "/api/auth/password/change" ||
		path == "/api/auth/account/disable")
}

func requiredPermission(method, path string) string {
	if method == http.MethodPost && path == "/api/v1/users" {
		return service.PermissionManageUsers
	}
	return ""
}

// UserID returns the authenticated user's database ID from the request context.
func UserID(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value(userIDContextKey).(uint)
	return userID, ok
}

func respondUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	utils.RespondUnauthorized(w, r, message)
}
