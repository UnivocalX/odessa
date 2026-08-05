package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Label struct {
	gorm.Model

	Name string `gorm:"type:varchar(64);not null;uniqueIndex" validate:"required,min=2,max=64"`
	Description string `gorm:"type:varchar(255);not null;default:''" validate:"max=255"`
}

// normalize fields
func (m *Label) BeforeSave(tx *gorm.DB) error {
	m.Name = strings.TrimSpace(strings.ToLower(m.Name))
	return nil
}

type BlobLabel struct {
	gorm.Model

	Label   Label  `gorm:"foreignKey:LabelID;constraint:OnDelete:CASCADE"`
	LabelID uint   `gorm:"not null;uniqueIndex:idx_blob_labels_label_blob" validate:"required"`
	BlobID  uint   `gorm:"not null;uniqueIndex:idx_blob_labels_label_blob" validate:"required"`
	Value   string `gorm:"not null;default:''"`
}

func (r *Repository) CreateLabel(ctx context.Context, name string, description string) (*Label, error) {
	label := &Label{
		Name:        name,
		Description: description,
	}

	if err := validate.Struct(label); err != nil {
		return nil, err
	}

	err := gorm.G[Label](r.DB).Create(ctx, label)
	if err != nil && isDuplicateKeyError(err) {
		return nil, ErrAlreadyExists
	}

	return label, err
}

func (r *Repository) ListLabels(ctx context.Context) ([]Label, error) {
	var labels []Label

	err := r.DB.WithContext(ctx).
		Find(&labels).
		Error

	return labels, err
}

func (r *Repository) GetLabel(ctx context.Context, id uint) (*Label, error) {
	var label Label

	err := r.DB.WithContext(ctx).
		First(&label, id).
		Error
	if err != nil {
		return nil, err
	}

	return &label, nil
}

func (r *Repository) GetLabelByName(ctx context.Context, name string) (*Label, error) {
	var label Label

	err := r.DB.WithContext(ctx).
		Where("name = ?", name).
		First(&label).
		Error
	if err != nil {
		return nil, err
	}

	return &label, nil
}

// AssignLabel attaches a label to a blob. If the label is already attached, it updates the value.
func (r *Repository) AssignLabel(ctx context.Context, blobID uint, labelID uint, value string) error {
	bl := BlobLabel{
		BlobID:  blobID,
		LabelID: labelID,
		Value:   value,
	}

	return r.DB.WithContext(ctx).
		Where("blob_id = ? AND label_id = ?", blobID, labelID).
		Assign(BlobLabel{Value: value}).
		FirstOrCreate(&bl).
		Error
}

// BatchLabelInput represents a single label assignment for batch processing.
type BatchLabelInput struct {
	BlobID  uint
	LabelID uint
	Value   string
}

// BatchLabelResult holds outcomes of a batch label assignment.
type BatchLabelResult struct {
	Assigned int
	Failed   int
	Errors   []string
}

// BatchAssignLabels upserts label assignments in a single transaction.
// Returns an error for any label that doesn't exist.
func (r *Repository) BatchAssignLabels(ctx context.Context, inputs []BatchLabelInput) (*BatchLabelResult, error) {
	result := &BatchLabelResult{}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, in := range inputs {
			var label Label
			if err := tx.Where("id = ?", in.LabelID).First(&label).Error; err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("label %d not found: %v", in.LabelID, err))
				continue
			}

			bl := BlobLabel{
				BlobID:  in.BlobID,
				LabelID: in.LabelID,
				Value:   in.Value,
			}

			if err := tx.
				Where("blob_id = ? AND label_id = ?", in.BlobID, in.LabelID).
				Assign(BlobLabel{Value: in.Value}).
				FirstOrCreate(&bl).Error; err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("blob %d label %d: %v", in.BlobID, in.LabelID, err))
				continue
			}
			result.Assigned++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repository: batch assign labels: %w", err)
	}

	return result, nil
}
