package repository

import (
	"context"
	"fmt"

	"encoding/json"

	"github.com/UnivocalX/odessa/internal/storage"
	"gorm.io/gorm"
)

// Origin represents a storage root or prefix
// (e.g. s3://bucket/prefix/, file:///data/).
type Origin struct {
	gorm.Model

	URI   storage.URI     `gorm:"not null;uniqueIndex" validate:"required,storage_uri"`
	Rules json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" validate:"json"`
}

func (r *Repository) CreateOrigin(ctx context.Context, uri string, rules json.RawMessage) (*Origin, error) {
	if rules == nil {
		rules = json.RawMessage(`{}`)
	}

	origin := &Origin{
		URI:   storage.URI(uri),
		Rules: rules,
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

// UpdateOriginRules replaces the label rules on an existing origin.
func (r *Repository) UpdateOriginRules(ctx context.Context, id uint, rules json.RawMessage) error {
	if rules == nil {
		rules = json.RawMessage(`{}`)
	}
	return r.DB.WithContext(ctx).
		Model(&Origin{}).
		Where("id = ?", id).
		Update("rules", rules).Error
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

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type ScanOrigin struct {
	gorm.Model

	OriginID uint            `gorm:"not null;index"  validate:"required"` // was: uniqueIndex
	Status   Status          `gorm:"type:text;not null;default:'pending'" validate:"required,oneof=pending in_progress completed failed cancelled"`
	Attempts int             `gorm:"not null;default:0"`
	Rules    json.RawMessage `gorm:"column:rules;type:jsonb;not null;default:'{}'" validate:"json"`
	Results  json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" validate:"required,json"`
}

// LabelRules maps glob patterns to label assignments.
// The key "*" applies to all files; any other key is a filepath.Match pattern.
//
// Example:
//
//	{
//	  "*": [{"label": "source", "value": "photos"}],
//	  "*.jpg": [{"label": "type", "value": "image"}]
//	}
type LabelRules map[string][]LabelAssignment

type LabelAssignment struct {
	Label string `json:"label" validate:"required"`
	Value string `json:"value"`
}

func (r *Repository) CreateScanOrigin(ctx context.Context, originID uint, rules json.RawMessage) (*ScanOrigin, error) {
	scan := &ScanOrigin{
		OriginID: originID,
		Status:   StatusPending,
		Rules:    rules,
		Results:  json.RawMessage(`{}`),
	}

	if err := r.DB.WithContext(ctx).Create(scan).Error; err != nil {
		if isUniqueViolation(err, "idx_scan_origins_active_origin") {
			return nil, fmt.Errorf("%w: origin %d", ErrScanAlreadyRunning, originID)
		}
		return nil, err
	}

	return scan, nil
}

func (r *Repository) GetScanOrigin(ctx context.Context, originID uint) (*ScanOrigin, error) {
	var scan ScanOrigin
	err := r.DB.WithContext(ctx).
		Where("origin_id = ?", originID).
		Last(&scan).Error
	if err != nil {
		return nil, fmt.Errorf("repository: get scan origin %d: %w", originID, err)
	}
	return &scan, nil
}

func (r *Repository) GetScanOriginByID(ctx context.Context, id uint) (*ScanOrigin, error) {
	var scan ScanOrigin
	err := r.DB.WithContext(ctx).
		Where("id = ?", id).
		First(&scan).Error
	if err != nil {
		return nil, fmt.Errorf("repository: get scan origin by id %d: %w", id, err)
	}
	return &scan, nil
}

func (r *Repository) ListScanOrigins(ctx context.Context) ([]ScanOrigin, error) {
	var scans []ScanOrigin
	if err := r.DB.WithContext(ctx).Order("created_at DESC").Find(&scans).Error; err != nil {
		return nil, fmt.Errorf("repository: list scan origins: %w", err)
	}
	return scans, nil
}

// ClaimScanOrigin atomically claims up to `limit` pending scan origins by
// transitioning them to in_progress within a transaction using
// SELECT ... FOR UPDATE SKIP LOCKED, ensuring no two workers process the same record.
func (r *Repository) ClaimScanOrigins(ctx context.Context, limit int) ([]ScanOrigin, error) {
	var scans []ScanOrigin

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Raw(`SELECT * FROM scan_origins
				 WHERE status = ? AND deleted_at IS NULL
				 ORDER BY created_at ASC
				 LIMIT ?
				 FOR UPDATE SKIP LOCKED`, StatusPending, limit).
			Scan(&scans).Error; err != nil {
			return err
		}

		if len(scans) == 0 {
			return nil
		}

		ids := make([]uint, len(scans))
		for i, s := range scans {
			ids[i] = s.ID
		}

		if err := tx.
			Model(&ScanOrigin{}).
			Where("id IN ?", ids).
			Update("status", StatusInProgress).Error; err != nil {
			return err
		}

		// Update the returned structs to reflect the new status.
		for i := range scans {
			scans[i].Status = StatusInProgress
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repository: claim scan origins: %w", err)
	}

	return scans, nil
}

// FailScanOrigin increments the attempt counter. If maxAttempts is reached,
// marks the scan as failed. Otherwise, re-enqueues it as pending for retry.
func (r *Repository) FailScanOrigin(ctx context.Context, id uint, maxAttempts int) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var scan ScanOrigin
		if err := tx.Where("id = ? AND status = ?", id, StatusInProgress).First(&scan).Error; err != nil {
			return fmt.Errorf("repository: fail scan origin %d: %w", id, err)
		}

		scan.Attempts++
		if scan.Attempts >= maxAttempts {
			scan.Status = StatusFailed
		} else {
			scan.Status = StatusPending
		}

		if err := tx.Save(&scan).Error; err != nil {
			return fmt.Errorf("repository: fail scan origin %d: %w", id, err)
		}
		return nil
	})
}

// CompleteScanOrigin marks a scan as completed with results.
func (r *Repository) CompleteScanOrigin(ctx context.Context, id uint, results any) error {
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("repository: marshal scan results: %w", err)
	}

	if err := r.DB.WithContext(ctx).
		Model(&ScanOrigin{}).
		Where("id = ? AND status = ?", id, StatusInProgress).
		Updates(map[string]any{
			"status":  StatusCompleted,
			"results": resultsJSON,
		}).Error; err != nil {
		return fmt.Errorf("repository: complete scan origin %d: %w", id, err)
	}
	return nil
}

// UpdateScanOriginStatus updates the status field for a scan origin record.
func (r *Repository) UpdateScanOriginStatus(ctx context.Context, id uint, status Status) error {
	return r.DB.WithContext(ctx).
		Model(&ScanOrigin{}).
		Where("id = ?", id).
		Update("status", status).Error
}
