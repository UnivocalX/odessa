package dto

import "github.com/UnivocalX/odessa/internal/repository"

type PostOriginRequest struct {
	URI string `json:"uri" validate:"required,url"`
}

type PostBlobRequest struct {
	URI string `json:"uri" validate:"required,url"`
}

type CreateUserRequest struct {
	Name     string            `json:"name" validate:"required,min=2,max=100"`
	Email    string            `json:"email" validate:"required,email"`
	Password repository.Secret `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string            `json:"email" validate:"required,email"`
	Password repository.Secret `json:"password" validate:"required,min=8,max=72"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword repository.Secret `json:"current_password" validate:"required,min=8,max=72"`
	NewPassword     repository.Secret `json:"new_password" validate:"required,min=8,max=72"`
}

type DisableAccountRequest struct {
	Password repository.Secret `json:"password" validate:"required,min=8,max=72"`
}

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PasswordResetConfirmRequest struct {
	Token    string            `json:"token" validate:"required"`
	Password repository.Secret `json:"password" validate:"required,min=8,max=72"`
}
