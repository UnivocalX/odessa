package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthOptions struct {
	JWTSecret            repository.Secret
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration
	ResetTokenLifetime   time.Duration
	PasswordResetURL     string
	EmailSender          EmailSender
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// ValidateAccessToken validates an access token and checks current user state.
func (s *Service) ValidateAccessToken(ctx context.Context, rawToken string) (uint, error) {
	claims := &accessClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.auth.JWTSecret.Expose()), nil
	}, jwt.WithIssuer("odessa"), jwt.WithAudience("odessa-api"))
	if err != nil || !token.Valid {
		return 0, ErrInvalidCredentials
	}

	userID, err := strconv.ParseUint(claims.Subject, 10, 0)
	if err != nil || userID == 0 {
		return 0, ErrInvalidCredentials
	}
	user, err := s.repo.GetUser(ctx, uint(userID))
	if err != nil || user.DisabledAt != nil || claims.TokenVersion != user.TokenVersion {
		return 0, ErrInvalidCredentials
	}
	return uint(userID), nil
}

// NewUser validates and hashes a plaintext password before persisting it.
func (s *Service) NewUser(ctx context.Context, name, email string, password repository.Secret) (*repository.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password.Expose()), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, name, email, repository.Secret(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, fmt.Errorf("%w: user", ErrAlreadyExists)
		}

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return nil, fmt.Errorf("%w: %w", ErrValidation, err)
		}
		return nil, err
	}

	return user, nil
}

// Login verifies a user's credentials and returns a signed access token.
func (s *Service) Login(ctx context.Context, email string, password repository.Secret) (*AuthTokens, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.DisabledAt != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password.Expose()),
		[]byte(password.Expose()),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	refresh, err := newOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}
	refreshSession := &repository.RefreshSession{
		UserID: user.ID, TokenHash: hashToken(refresh),
		ExpiresAt: now.Add(s.auth.RefreshTokenLifetime),
	}
	if err := s.repo.CreateRefreshSession(ctx, refreshSession); err != nil {
		return nil, fmt.Errorf("persist refresh session: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "odessa",
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			Audience:  jwt.ClaimStrings{"odessa-api"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.auth.AccessTokenLifetime)),
		},
		TokenVersion: user.TokenVersion,
	})

	signed, err := token.SignedString([]byte(s.auth.JWTSecret.Expose()))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	return &AuthTokens{AccessToken: signed, RefreshToken: refresh, ExpiresIn: int(s.auth.AccessTokenLifetime.Seconds())}, nil
}

func (s *Service) RefreshSession(ctx context.Context, refresh string) (*AuthTokens, error) {
	session, err := s.repo.GetRefreshSession(ctx, hashToken(refresh))
	if err != nil || session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return nil, ErrInvalidCredentials
	}
	user, err := s.repo.GetUser(ctx, session.UserID)
	if err != nil || user.DisabledAt != nil {
		return nil, ErrInvalidCredentials
	}

	next, err := newOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}
	now := time.Now()
	session.RevokedAt = &now
	session.ReplacedBy = hashToken(next)
	if err := s.repo.UpdateRefreshSession(ctx, session); err != nil {
		return nil, fmt.Errorf("rotate refresh token: %w", err)
	}
	if err := s.repo.CreateRefreshSession(ctx, &repository.RefreshSession{
		UserID: session.UserID, TokenHash: hashToken(next),
		ExpiresAt: now.Add(s.auth.RefreshTokenLifetime),
	}); err != nil {
		return nil, fmt.Errorf("persist rotated refresh token: %w", err)
	}
	return s.issueAccessToken(ctx, session.UserID, next, now)
}

func (s *Service) RevokeRefreshToken(ctx context.Context, refresh string) error {
	session, err := s.repo.GetRefreshSession(ctx, hashToken(refresh))
	if err != nil {
		return nil
	}
	now := time.Now()
	session.RevokedAt = &now
	return s.repo.UpdateRefreshSession(ctx, session)
}

// Logout revokes the refresh session and invalidates all access tokens for the user.
func (s *Service) Logout(ctx context.Context, userID uint, refresh string) error {
	if err := s.RevokeRefreshToken(ctx, refresh); err != nil {
		return err
	}
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	user.TokenVersion++
	return s.repo.UpdateUser(ctx, user)
}

func (s *Service) ChangePassword(ctx context.Context, userID uint, current, next repository.Secret) error {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil || user.DisabledAt != nil {
		return ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.Expose()), []byte(current.Expose())); err != nil {
		return ErrInvalidCredentials
	}
	return s.setPassword(ctx, user, next)
}

func (s *Service) DisableAccount(ctx context.Context, userID uint, password repository.Secret) error {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil || user.DisabledAt != nil {
		return ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.Expose()), []byte(password.Expose())); err != nil {
		return ErrInvalidCredentials
	}
	now := time.Now()
	user.DisabledAt = &now
	user.TokenVersion++
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.repo.RevokeRefreshSessions(ctx, userID)
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || user.DisabledAt != nil {
		return nil
	}
	token, err := newOpaqueToken()
	if err != nil {
		return err
	}
	if err := s.repo.CreatePasswordResetToken(ctx, &repository.PasswordResetToken{
		UserID: user.ID, TokenHash: hashToken(token), ExpiresAt: time.Now().Add(s.auth.ResetTokenLifetime),
	}); err != nil {
		return err
	}
	if s.auth.EmailSender == nil {
		return nil
	}
	resetLink := strings.TrimRight(s.auth.PasswordResetURL, "/") + "?token=" + token
	return s.auth.EmailSender.SendPasswordReset(ctx, PasswordResetEmail{
		To: user.Email, Name: user.Name, ResetLink: resetLink,
	})
}

func (s *Service) ResetPassword(ctx context.Context, token string, password repository.Secret) error {
	reset, err := s.repo.GetPasswordResetToken(ctx, hashToken(token))
	if err != nil || reset.UsedAt != nil || time.Now().After(reset.ExpiresAt) {
		return ErrInvalidCredentials
	}
	user, err := s.repo.GetUser(ctx, reset.UserID)
	if err != nil || user.DisabledAt != nil {
		return ErrInvalidCredentials
	}
	if err := s.setPassword(ctx, user, password); err != nil {
		return err
	}
	now := time.Now()
	reset.UsedAt = &now
	if err := s.repo.UpdatePasswordResetToken(ctx, reset); err != nil {
		return err
	}
	return s.repo.RevokeRefreshSessions(ctx, user.ID)
}

func (s *Service) setPassword(ctx context.Context, user *repository.User, password repository.Secret) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password.Expose()), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.Password = repository.Secret(hashed)
	user.TokenVersion++
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.repo.RevokeRefreshSessions(ctx, user.ID)
}

func (s *Service) issueAccessToken(ctx context.Context, userID uint, refresh string, now time.Time) (*AuthTokens, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "odessa", Subject: strconv.FormatUint(uint64(userID), 10), Audience: jwt.ClaimStrings{"odessa-api"},
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(s.auth.AccessTokenLifetime)),
		},
		TokenVersion: user.TokenVersion,
	})
	signed, err := token.SignedString([]byte(s.auth.JWTSecret.Expose()))
	if err != nil {
		return nil, err
	}
	return &AuthTokens{AccessToken: signed, RefreshToken: refresh, ExpiresIn: int(s.auth.AccessTokenLifetime.Seconds())}, nil
}

type accessClaims struct {
	jwt.RegisteredClaims
	TokenVersion int `json:"token_version"`
}

func newOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
