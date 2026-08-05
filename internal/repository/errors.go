package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAlreadyExists      = errors.New("record already exists")
	ErrNotFound           = errors.New("record not found")
	ErrScanAlreadyRunning = errors.New("scan already running for origin")
)

// MultiError represents multiple keyed errors.
// It supports errors.Is/errors.As via the Go 1.20+ multi-unwrap interface.
type MultiError struct {
	Errs map[string]error
}

func (e *MultiError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d error(s):", len(e.Errs))
	for key, err := range e.Errs {
		fmt.Fprintf(&b, "\n  %s: %s", key, err)
	}
	return b.String()
}

func (e *MultiError) Unwrap() []error {
	errList := make([]error, 0, len(e.Errs))
	for _, err := range e.Errs {
		errList = append(errList, err)
	}
	return errList
}

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

func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
	}
	return false
}
