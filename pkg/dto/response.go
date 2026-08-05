package dto

import (
	"encoding/json"
	"time"
)

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
	ID        uint            `json:"id"`
	URI       string          `json:"uri"`
	CreatedAt string          `json:"created_at"`
	Rules     json.RawMessage `json:"rules,omitempty"`
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
	OriginID  uint   `json:"origin_id"`
	Status    string `json:"status"`
	Attempts  int    `json:"attempts"`
	CreatedAt string `json:"created_at"`
}

type LabelResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"Name"`
	Description string `json:"description"`
}

type ListLabelsResponse struct {
	Labels []LabelResponse `json:"labels"`
}

type BlobResponse struct {
	ID        uint                `json:"id"`
	Hash      string              `json:"hash"`
	MimeType  string              `json:"mime_type"`
	Size      int64               `json:"size"`
	Labels    []BlobLabelResponse `json:"labels,omitempty"`
	Locations []string            `json:"locations,omitempty"`
}

type LocationResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BlobLabelResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SearchBlobsResponse struct {
	Blobs      []BlobResponse `json:"blobs"`
	NextCursor uint           `json:"next_cursor"`
	HasMore    bool           `json:"has_more"`
}
