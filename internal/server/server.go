package server

import (
	"net/http"
	"time"

	"github.com/UnivocalX/odessa/internal/server/api"
	"github.com/UnivocalX/odessa/internal/service"
)

type Config struct {
	Addr                string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	MaxHeaderBytes      int
	MaxRequestBodyBytes int64
}

func New(authSvc *service.AuthService, blobSvc *service.BlobService, cfg Config) *http.Server {
	mux := http.NewServeMux()
	handler := api.Register(mux, authSvc, blobSvc, cfg.MaxRequestBodyBytes)

	return &http.Server{
		Addr:           cfg.Addr,
		Handler:        handler,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		IdleTimeout:    cfg.IdleTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}
}
