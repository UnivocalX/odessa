package server

import (
	"net/http"
	"time"

	"example.com/aether/internal/server/api"
	"example.com/aether/internal/repository"
	"example.com/aether/internal/service"
	"example.com/aether/internal/storage"
)

func New(addr string, repo *repository.Repository, reg *storage.Registry) *http.Server {
	svc := service.New(repo, reg)

	mux := http.NewServeMux()
	handler := api.Register(mux, svc)

	return &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
