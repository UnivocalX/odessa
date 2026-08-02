package service

import (
	"time"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/storage"
)

// AuthService handles authentication, user management, and authorization.
type AuthService struct {
	repo *repository.Repository
	auth AuthOptions
}

func NewAuthService(repo *repository.Repository, authOptions AuthOptions) *AuthService {
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
	auth.PasswordResetURL = authOptions.PasswordResetURL
	auth.EmailSender = authOptions.EmailSender
	return &AuthService{repo: repo, auth: auth}
}

// BlobService handles origins, scans, and blob operations.
type BlobService struct {
	repo *repository.Repository
	reg  *storage.Registry
}

func NewBlobService(repo *repository.Repository, reg *storage.Registry) *BlobService {
	return &BlobService{repo: repo, reg: reg}
}
