package dto

type PostOriginRequest struct {
	URI string `json:"uri" validate:"required,url"`
}

type PostBlobRequest struct {
	URI string `json:"uri" validate:"required,url"`
}