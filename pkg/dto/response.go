package dto

import "time"

type Response struct {
	Message string `json:"message,omitempty"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Response
	Error string `json:"error,omitempty"`
}

type OriginResponse struct {
	ID        uint   `json:"id"`
	URI       string `json:"uri"`
	CreatedAt string `json:"created_at"`
}

type ListOriginsResponse struct {
	Origins []OriginResponse `json:"origins"`
}

type UserResponse struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	DisabledAt *time.Time `json:"disabled_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
}

type ScanOriginResponse struct {
	ID        uint   `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
