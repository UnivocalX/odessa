package server

import (
	"net/http"
	"time"

	"github.com/UnivocalX/odessa/internal/server/api"
	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/internal/storage"
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
