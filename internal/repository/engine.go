package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

type Repository struct {
	DB       *gorm.DB
	listener *pgx.Conn
}

func New(cfg Config) (*Repository, error) {
	db, err := Open(cfg)
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		return nil, err
	}

	// Create a dedicated pgx connection for LISTEN/NOTIFY.
	ln, err := pgx.Connect(context.Background(), cfg.DSN.Expose())
	if err != nil {
		// close db if listener can't be created
		if sqlDB, dErr := db.DB(); dErr == nil {
			sqlDB.Close()
		}
		return nil, fmt.Errorf("repository: open notify connection: %w", err)
	}

	return &Repository{DB: db, listener: ln}, nil
}

// Close closes the repository's resources (listener and DB).
func (r *Repository) Close() error {
	var firstErr error
	if r == nil {
		return nil
	}
	if r.listener != nil {
		_ = r.listener.Close(context.Background())
	}
	if r.DB != nil {
		if sqlDB, err := r.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
