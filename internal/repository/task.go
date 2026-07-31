package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type TaskType string

const (
	TaskTypeScanOrigin TaskType = "scan_origin"
)

type Task struct {
	gorm.Model

	Type    TaskType        `gorm:"type:text;not null" validate:"required,oneof=scan_origin"`
	Status  TaskStatus      `gorm:"type:text;not null;default:'pending'" validate:"required,oneof=pending in_progress completed failed"`
	Details json.RawMessage `gorm:"type:jsonb;not null" validate:"required,json"`
}

type ScanOriginDetails struct {
	OriginID uint `json:"origin_id" validate:"required,gt=0"`
}

func (r *Repository) GetTask(ctx context.Context, id uint) (*Task, error) {
	var task Task
	if err := r.DB.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, fmt.Errorf("repository: get task %d: %w", id, err)
	}
	return &task, nil
}

func (r *Repository) HasActiveScanForOrigin(ctx context.Context, originID uint) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&Task{}).
		Where("type = ? AND status IN (?, ?) AND details->>'origin_id' = ?",
			TaskTypeScanOrigin, TaskStatusPending, TaskStatusInProgress, fmt.Sprintf("%d", originID)).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("repository: check active scan: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) CreateTask(
	ctx context.Context,
	taskType TaskType,
	details any,
) (*Task, error) {
	if err := validate.Struct(details); err != nil {
		return nil, fmt.Errorf("repository: validate task details: %w", err)
	}

	raw, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("repository: marshal task details: %w", err)
	}

	task := &Task{
		Type:    taskType,
		Status:  TaskStatusPending,
		Details: raw,
	}

	if err := validate.Struct(task); err != nil {
		return nil, fmt.Errorf("repository: validate task: %w", err)
	}

	if err := r.DB.WithContext(ctx).Create(task).Error; err != nil {
		return nil, fmt.Errorf("repository: create task: %w", err)
	}

	return task, nil
}