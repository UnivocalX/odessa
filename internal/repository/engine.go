package repository

import (
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func New(cfg Config) (*Repository, error) {
	db, err := Open(cfg)
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		return nil, err
	}
	return &Repository{DB: db}, nil
}
