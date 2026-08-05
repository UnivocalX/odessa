package repository

import (
	"context"
	"fmt"

	"github.com/UnivocalX/odessa/internal/core"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Blob struct {
	gorm.Model

	Hash      string      `gorm:"not null;uniqueIndex" validate:"required"`
	MimeType  string      `gorm:"not null;default:''" validate:"required"`
	Size      int64       `gorm:"not null;default:0" validate:"gte=0"`
	Locations []Location  `gorm:"foreignKey:BlobID;constraint:OnDelete:CASCADE" validate:"dive"`
	Labels    []BlobLabel `gorm:"foreignKey:BlobID;constraint:OnDelete:CASCADE" validate:"dive"`
}

type Location struct {
	gorm.Model

	BlobID uint     `gorm:"not null;index" validate:"required"`
	URI    core.URI `gorm:"not null;uniqueIndex:idx_locations_uri" validate:"required,url"`
}

// BlobBatchCreateInput represents a discovered file to be persisted.
type BlobBatchCreateInput struct {
	Hash     string
	Size     int64
	MimeType string
	URI      string
}

// BlobBatchCreateResult holds per-item outcomes of a batch create.
type BlobBatchCreateResult struct {
	Created int
	Failed  int
	Errors  []string
}

// BatchCreateBlobs upserts blobs and their locations in a single transaction.
// For each input, if a blob with the same hash already exists, it adds a new
// location pointing to it; otherwise it creates both the blob and the location.
func (r *Repository) BatchCreateBlobs(ctx context.Context, inputs []BlobBatchCreateInput) (*BlobBatchCreateResult, error) {
	result := &BlobBatchCreateResult{}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, in := range inputs {
			blob := Blob{
				Hash:     in.Hash,
				MimeType: in.MimeType,
				Size:     in.Size,
			}

			// Upsert: on hash conflict, update updated_at so we get the ID back.
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "hash"}},
				DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
			}).Create(&blob).Error; err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("%s: create blob: %v", in.URI, err))
				continue
			}

			loc := Location{
				BlobID: blob.ID,
				URI:    core.URI(in.URI),
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "uri"}},
				DoUpdates: clause.AssignmentColumns([]string{"blob_id", "updated_at"}),
			}).Create(&loc).Error; err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("%s: create location: %v", in.URI, err))
				continue
			}

			result.Created++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repository: batch create blobs: %w", err)
	}

	return result, nil
}

func (r *Repository) GetBlobByHash(ctx context.Context, hash string) (*Blob, error) {
	var blob Blob
	if err := r.DB.WithContext(ctx).
		Preload("Locations").
		Preload("Labels").
		Preload("Labels.Label").
		Where("hash = ?", hash).
		First(&blob).Error; err != nil {
		return nil, err
	}
	return &blob, nil
}

func (r *Repository) GetBlobByID(ctx context.Context, id uint) (*Blob, error) {
	var blob Blob
	if err := r.DB.WithContext(ctx).
		Preload("Locations").
		Preload("Labels").
		Preload("Labels.Label").
		First(&blob, id).Error; err != nil {
		return nil, err
	}
	return &blob, nil
}
