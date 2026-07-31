package service

import (
	"time"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/storage"
)

type Service struct {
	repo *repository.Repository
	reg  *storage.Registry
	auth AuthOptions
}

func New(repo *repository.Repository, reg *storage.Registry, authOptions AuthOptions) *Service {
	auth := AuthOptions{AccessTokenLifetime: 15 * time.Minute, RefreshTokenLifetime: 30 * 24 * time.Hour, ResetTokenLifetime: time.Hour}
	auth.JWTSecret = authOptions.JWTSecret
	if authOptions.AccessTokenLifetime > 0 {
		auth.AccessTokenLifetime = authOptions.AccessTokenLifetime
	}
	if authOptions.RefreshTokenLifetime > 0 {
		auth.RefreshTokenLifetime = authOptions.RefreshTokenLifetime
	}
	if authOptions.ResetTokenLifetime > 0 {
		auth.ResetTokenLifetime = authOptions.ResetTokenLifetime
	}
	return &Service{repo: repo, reg: reg, auth: auth}
}
