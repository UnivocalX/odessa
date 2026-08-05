package repository

import (
	"errors"

	"github.com/UnivocalX/odessa/internal/core"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrScanAlreadyRunning = errors.New("scan already running for origin")
)

type MultiError = core.MultiError

// isDuplicateKeyError returns true if err indicates a unique constraint violation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
	}
	return false
}

func isForeignKeyViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503" && pgErr.ConstraintName == constraintName
	}
	return false
}
