package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type RefreshSession struct {
	gorm.Model
	UserID     uint      `gorm:"not null;index"`
	TokenHash  string    `gorm:"not null;uniqueIndex"`
	ExpiresAt  time.Time `gorm:"not null;index"`
	RevokedAt  *time.Time
	ReplacedBy string
}

func (r *Repository) CreateRefreshSession(ctx context.Context, session *RefreshSession) error {
	return r.DB.WithContext(ctx).Create(session).Error
}

func (r *Repository) GetRefreshSession(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	var session RefreshSession
	if err := r.DB.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) UpdateRefreshSession(ctx context.Context, session *RefreshSession) error {
	return r.DB.WithContext(ctx).Save(session).Error
}

func (r *Repository) RevokeRefreshSessions(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.DB.WithContext(ctx).Model(&RefreshSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

type PasswordResetToken struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	TokenHash string    `gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time
}

func (r *Repository) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	return r.DB.WithContext(ctx).Create(token).Error
}

func (r *Repository) GetPasswordResetToken(ctx context.Context, hash string) (*PasswordResetToken, error) {
	var token PasswordResetToken
	if err := r.DB.WithContext(ctx).Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *Repository) UpdatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	return r.DB.WithContext(ctx).Save(token).Error
}
