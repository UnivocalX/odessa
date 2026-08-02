package server

import (
	"net/http"
	"time"

	"github.com/UnivocalX/odessa/internal/server/api"
	"github.com/UnivocalX/odessa/internal/service"
)

type Options struct {
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	MaxHeaderBytes      int
	MaxRequestBodyBytes int64
}

type Config struct {
	Addr string
	HTTP Options
}

func New(authSvc *service.AuthService, blobSvc *service.BlobService, cfg Config) *http.Server {
	mux := http.NewServeMux()
	handler := api.Register(mux, authSvc, blobSvc, cfg.HTTP.MaxRequestBodyBytes)

	return &http.Server{
		Addr:           cfg.Addr,
		Handler:        handler,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		IdleTimeout:    cfg.HTTP.IdleTimeout,
		MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
	}
}
