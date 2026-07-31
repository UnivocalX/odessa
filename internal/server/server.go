package server

import (
	"net/http"
	"time"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/server/api"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/internal/storage"
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
	Auth service.AuthOptions
}

func New(repo *repository.Repository, reg *storage.Registry, cfg Config) *http.Server {
	svc := service.New(repo, reg, cfg.Auth)

	mux := http.NewServeMux()
	handler := api.Register(mux, svc, cfg.HTTP.MaxRequestBodyBytes)

	return &http.Server{
		Addr:           cfg.Addr,
		Handler:        handler,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		IdleTimeout:    cfg.HTTP.IdleTimeout,
		MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
	}
}
