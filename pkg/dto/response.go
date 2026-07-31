package dto

type Response struct {
	Message string `json:"message,omitempty"`
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

type TaskResponse struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
