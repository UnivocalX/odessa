package dto

import (
	"github.com/UnivocalX/odessa/internal/core"
	"github.com/UnivocalX/odessa/internal/repository"
)

type PostOriginRequest struct {
	URI   string                 `json:"uri" validate:"required,url"`
	Rules *repository.LabelRules `json:"rules,omitempty"`
}

type PostBlobRequest struct {
	URI string `json:"uri" validate:"required,url"`
}

type PostLabelRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=64"`
	Description string `json:"description" validate:"max=255"`
}

type PostScanOriginRequest struct {
	Rules *repository.LabelRules `json:"rules,omitempty"`
}

type PutOriginRulesRequest struct {
	Rules repository.LabelRules `json:"rules" validate:"required"`
}

type CreateUserRequest struct {
	Name     string      `json:"name" validate:"required,min=2,max=100"`
	Email    string      `json:"email" validate:"required,email"`
	Password core.Secret `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string      `json:"email" validate:"required,email"`
	Password core.Secret `json:"password" validate:"required,min=8,max=72"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword core.Secret `json:"current_password" validate:"required,min=8,max=72"`
	NewPassword     core.Secret `json:"new_password" validate:"required,min=8,max=72"`
}

type DisableAccountRequest struct {
	Password core.Secret `json:"password" validate:"required,min=8,max=72"`
}

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PasswordResetConfirmRequest struct {
	Token    string      `json:"token" validate:"required"`
	Password core.Secret `json:"password" validate:"required,min=8,max=72"`
}

type BlobFilter struct {
	Hashes      []string          `json:"hashes,omitempty"`
	MimeTypes   []string          `json:"mime_types,omitempty"`
	Labels      []string          `json:"labels,omitempty"`
	LabelValues map[string]string `json:"label_values,omitempty"`
	URIPattern  string            `json:"uri_pattern,omitempty"`
}

type SearchBlobsRequest struct {
	Include BlobFilter `json:"include,omitempty"`
	Exclude BlobFilter `json:"exclude,omitempty"`

	MinSize *int64 `json:"min_size,omitempty"`
	MaxSize *int64 `json:"max_size,omitempty"`

	Cursor uint `json:"cursor,omitempty"`
	Limit  int  `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

type PostDatasetRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=64"`
	Description string `json:"description" validate:"max=255"`
}

type PostDatasetVersionRequest struct {
	Commit  string `json:"commit" validate:"max=255"`
	BlobIDs []uint `json:"blob_ids" validate:"required,min=1,dive,min=1"`
}
