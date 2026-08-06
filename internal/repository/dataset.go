package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/UnivocalX/odessa/internal/core"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Dataset struct {
	gorm.Model

	Name        string           `gorm:"type:varchar(64);not null;uniqueIndex" validate:"required,min=2,max=64"`
	Description string           `gorm:"type:varchar(255);not null;default:''" validate:"max=255"`
	Versions    []DatasetVersion `gorm:"foreignKey:DatasetID;constraint:OnDelete:CASCADE" validate:"dive"`
}

// normalize fields
func (m *Dataset) BeforeSave(tx *gorm.DB) error {
	m.Name = strings.TrimSpace(strings.ToLower(m.Name))
	return nil
}

type DatasetVersion struct {
	gorm.Model

	Commit    string `gorm:"type:varchar(255);not null;default:''" validate:"max=255"`
	DatasetID uint   `gorm:"not null;index" validate:"required"`
	Blobs     []Blob `gorm:"many2many:dataset_version_blobs;constraint:OnDelete:CASCADE" validate:"dive"`
}

type DatasetVersionBlob struct {
	DatasetVersionID uint `gorm:"primaryKey"`
	BlobID           uint `gorm:"primaryKey"`
}

func (DatasetVersionBlob) TableName() string {
	return "dataset_version_blobs"
}

func (r *Repository) CreateDataset(ctx context.Context, name string, description string) (*Dataset, error) {
	dataset := &Dataset{
		Name:        name,
		Description: description,
	}

	if err := validate.Struct(dataset); err != nil {
		return nil, err
	}

	err := gorm.G[Dataset](r.DB).Create(ctx, dataset)
	if err != nil && isDuplicateKeyError(err) {
		return nil, core.ErrAlreadyExists
	}

	return dataset, err
}

func (r *Repository) ListDatasets(ctx context.Context) ([]Dataset, error) {
	var datasets []Dataset

	if err := r.DB.WithContext(ctx).
		Order("created_at DESC").
		Find(&datasets).Error; err != nil {
		return nil, err
	}

	return datasets, nil
}

func (r *Repository) CreateDatasetVersion(ctx context.Context, id uint, commit string) (*DatasetVersion, error) {
	version := &DatasetVersion{
		DatasetID: id,
		Commit:    commit,
	}

	if err := validate.Struct(version); err != nil {
		return nil, err
	}

	err := gorm.G[DatasetVersion](r.DB).Create(ctx, version)
	if isForeignKeyViolation(err, "dataset_versions_dataset_id_fkey") {
		return nil, core.ErrNotFound
	}

	return version, err
}

func (r *Repository) ListDatasetVersions(ctx context.Context, datasetID uint) ([]DatasetVersion, error) {
	var dataset Dataset
	if err := r.DB.WithContext(ctx).Select("id").First(&dataset, datasetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	var versions []DatasetVersion
	if err := r.DB.WithContext(ctx).
		Where("dataset_id = ?", datasetID).
		Order("created_at DESC").
		Find(&versions).Error; err != nil {
		return nil, err
	}

	return versions, nil
}

func (r *Repository) GetDatasetVersion(ctx context.Context, datasetID uint, versionID uint) (*DatasetVersion, error) {
	var version DatasetVersion
	if err := r.DB.WithContext(ctx).
		Where("id = ? AND dataset_id = ?", versionID, datasetID).
		First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	return &version, nil
}

// GetDatasetVersionBlobs returns a paginated list of blobs for a dataset version.
func (r *Repository) GetDatasetVersionBlobs(ctx context.Context, datasetID uint, versionID uint, cursor uint, limit int) (*BlobSearchResult, error) {
	var version DatasetVersion
	if err := r.DB.WithContext(ctx).
		Select("id").
		Where("id = ? AND dataset_id = ?", versionID, datasetID).
		First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := r.DB.WithContext(ctx).Model(&Blob{}).
		Joins("JOIN dataset_version_blobs ON dataset_version_blobs.blob_id = blobs.id").
		Where("dataset_version_blobs.dataset_version_id = ?", versionID)

	if cursor > 0 {
		query = query.Where("blobs.id > ?", cursor)
	}

	var blobs []Blob
	if err := query.Order("blobs.id ASC").Limit(limit + 1).
		Preload("Locations").Preload("Labels").Preload("Labels.Label").
		Find(&blobs).Error; err != nil {
		return nil, fmt.Errorf("repository: get dataset version blobs: %w", err)
	}

	result := &BlobSearchResult{}
	if len(blobs) > limit {
		result.HasMore = true
		blobs = blobs[:limit]
	}
	result.Blobs = blobs
	if len(blobs) > 0 {
		result.NextCursor = blobs[len(blobs)-1].ID
	}
	return result, nil
}

// BatchAssociateBlobsToDatasetVersion links many blobs to a dataset version.
// It is optimized for large inputs by deduplicating IDs and inserting in chunks.
func (r *Repository) BatchAssociateBlobsToDatasetVersion(ctx context.Context, datasetVersionID uint, blobIDs []uint) (int, error) {
	const chunkSize = 1000

	if len(blobIDs) == 0 {
		return 0, nil
	}

	associated := 0
	itemErrs := make(map[string]error)

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureDatasetVersionExists(tx, datasetVersionID); err != nil {
			return err
		}

		validIDs, errs := validateIDs(blobIDs)
		mergeErrors(itemErrs, errs)
		if len(validIDs) == 0 {
			return nil
		}

		found, err := findExistingBlobIDs(tx, validIDs, chunkSize)
		if err != nil {
			return err
		}

		links, errs := buildDatasetVersionBlobLinks(datasetVersionID, validIDs, found)
		mergeErrors(itemErrs, errs)
		if len(links) == 0 {
			return nil
		}

		inserted, err := insertDatasetVersionBlobLinks(tx, links, chunkSize)
		if err != nil {
			return err
		}
		associated += inserted

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("repository: batch associate blobs to dataset version: %w", err)
	}
	if len(itemErrs) > 0 {
		return associated, &MultiError{Errs: itemErrs}
	}

	return associated, nil
}

func ensureDatasetVersionExists(tx *gorm.DB, datasetVersionID uint) error {
	var version DatasetVersion
	if err := tx.Select("id").First(&version, datasetVersionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return core.ErrNotFound
		}
		return err
	}
	return nil
}

func validateIDs(ids []uint) ([]uint, map[string]error) {
	uniq := make(map[uint]struct{}, len(ids))
	validIDs := make([]uint, 0, len(ids))
	errs := make(map[string]error)
	for i, id := range ids {
		if id == 0 {
			errs[fmt.Sprintf("input[%d]", i)] = fmt.Errorf("blob id 0 is invalid")
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		validIDs = append(validIDs, id)
	}
	return validIDs, errs
}

func findExistingBlobIDs(tx *gorm.DB, ids []uint, chunkSize int) (map[uint]struct{}, error) {
	found := make(map[uint]struct{}, len(ids))
	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}

		var chunkIDs []uint
		if err := tx.Model(&Blob{}).
			Select("id").
			Where("id IN ?", ids[i:end]).
			Find(&chunkIDs).Error; err != nil {
			return nil, err
		}
		for _, id := range chunkIDs {
			found[id] = struct{}{}
		}
	}

	return found, nil
}

func buildDatasetVersionBlobLinks(datasetVersionID uint, ids []uint, found map[uint]struct{}) ([]DatasetVersionBlob, map[string]error) {
	links := make([]DatasetVersionBlob, 0, len(found))
	errs := make(map[string]error)
	for _, id := range ids {
		if _, ok := found[id]; !ok {
			errs[fmt.Sprintf("blob:%d", id)] = fmt.Errorf("blob %d not found", id)
			continue
		}
		links = append(links, DatasetVersionBlob{
			DatasetVersionID: datasetVersionID,
			BlobID:           id,
		})
	}
	return links, errs
}

func mergeErrors(dst map[string]error, src map[string]error) {
	for k, err := range src {
		dst[k] = err
	}
}

func insertDatasetVersionBlobLinks(tx *gorm.DB, links []DatasetVersionBlob, chunkSize int) (int, error) {
	associated := 0

	for i := 0; i < len(links); i += chunkSize {
		end := i + chunkSize
		if end > len(links) {
			end = len(links)
		}

		res := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "dataset_version_id"}, {Name: "blob_id"}},
			DoNothing: true,
		}).Create(links[i:end])
		if res.Error != nil {
			return 0, res.Error
		}
		associated += int(res.RowsAffected)
	}

	return associated, nil
}
