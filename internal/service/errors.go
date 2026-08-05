package service

import (
	"errors"
	"fmt"

	"github.com/UnivocalX/odessa/internal/core"
	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/go-playground/validator/v10"
)

var (
	// ErrUnsupportedBackend is returned when the URI scheme has no registered storage backend.
	ErrUnsupportedBackend = errors.New("unsupported storage backend")

	// ErrLocationNotFound is returned when the location does not exist.
	ErrLocationNotFound = errors.New("location not found")

	// ErrLocationNotAccessible is returned when the location exists but cannot be reached
	// (e.g. permission denied, invalid credentials for the bucket, etc.).
	ErrLocationNotAccessible = errors.New("location is not accessible")

	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrCannotCancel is returned when a scan cannot be cancelled (already finished).
	ErrCannotCancel = errors.New("cannot cancel scan")
)

func mapValidationError(err error) error {
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return fmt.Errorf("%w: %w", core.ErrValidation, err)
	}

	var itemErrs *repository.MultiError
	if errors.As(err, &itemErrs) {
		return fmt.Errorf("%w: %w", core.ErrValidation, err)
	}

	return nil
}
