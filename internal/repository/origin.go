package repository

import (
	"context"

	"github.com/UnivocalX/odessa/internal/storage"
	"gorm.io/gorm"
)

// Origin represents a storage root or prefix
// (e.g. s3://bucket/prefix/, file:///data/).
type Origin struct {
	gorm.Model

	URI storage.URI `gorm:"not null;uniqueIndex" validate:"required,storage_uri"`
}

func (r *Repository) CreateOrigin(ctx context.Context, uri string) (*Origin, error) {
	origin := &Origin{
		URI: storage.URI(uri),
	}
	
	if err := validate.Struct(origin); err != nil {
		return nil, err
	}

	err := gorm.G[Origin](r.DB).Create(ctx, origin)
	if err != nil && isDuplicateKeyError(err) {
		return nil, ErrAlreadyExists
	}

	return origin, err
}

func (r *Repository) ListOrigins(ctx context.Context) ([]Origin, error) {
	var origins []Origin

	err := r.DB.WithContext(ctx).
		Find(&origins).
		Error

	return origins, err
}

func (r *Repository) GetOrigin(ctx context.Context, id uint) (*Origin, error) {
	var origin Origin

	err := r.DB.WithContext(ctx).
		First(&origin, id).
		Error
	if err != nil {
		return nil, err
	}

	return &origin, nil
}