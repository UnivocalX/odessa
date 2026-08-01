package service

import "errors"

var (
	ErrValidation = errors.New("validation failed")

	// ErrUnsupportedBackend is returned when the URI scheme has no registered storage backend.
	ErrUnsupportedBackend = errors.New("unsupported storage backend")

	// ErrLocationNotFound is returned when the location does not exist.
	ErrLocationNotFound = errors.New("location not found")

	// ErrLocationNotAccessible is returned when the location exists but cannot be reached
	// (e.g. permission denied, invalid credentials for the bucket, etc.).
	ErrLocationNotAccessible = errors.New("location is not accessible")

	// ErrAlreadyExists is returned when the resource already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrNotFound is returned when the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
