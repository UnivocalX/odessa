package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAlreadyExists = errors.New("record already exists")
	ErrNotFound      = errors.New("record not found")
)

// isDuplicateKeyError returns true if err indicates a unique constraint violation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	// PostgreSQL: unique_violation (23505)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}

	return false
}
