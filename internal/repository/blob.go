package repository

import (
	"example.com/aether/internal/storage"
	"gorm.io/gorm"
)

type Blob struct {
	gorm.Model

	Hash      string     `gorm:"not null;uniqueIndex" validate:"required"`
	MimeType  string     `gorm:"not null;default:''" validate:"required"`
	Size      int64      `gorm:"not null;default:0" validate:"gte=0"`
	Locations []Location `gorm:"foreignKey:BlobID;constraint:OnDelete:CASCADE" validate:"dive"`
}

type Location struct {
	gorm.Model

	BlobID uint        `gorm:"not null;index"`
	URI    storage.URI `gorm:"not null" validate:"required,storage_uri"`
}
