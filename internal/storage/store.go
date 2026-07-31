package storage

import (
	"context"
	"io"
	"net/url"
)

// URI is a storage location routed to a backend by its scheme.
//
// Supported schemes:
//   - file:// – local filesystem (fs backend)
//   - s3://   – AWS S3 or S3-compatible storage
//   - az://   – Azure Blob Storage
type URI string

func (u URI) Parse() (*url.URL, error) { return url.Parse(string(u)) }

// Store is the backend-agnostic interface for object storage.
// All backends (local, S3, Azure) implement this interface so callers
// interact with a single API regardless of where data lives.
type Store interface {
	// Get opens the object at location for reading.
	// The caller must close the returned ReadCloser.
	Get(ctx context.Context, location string) (io.ReadCloser, error)

	// Put writes r to location, creating or overwriting the object.
	Put(ctx context.Context, location string, r io.Reader) error

	// Delete removes the object at location.
	Delete(ctx context.Context, location string) error

	// Available reports whether an object is accessible.
	Available(ctx context.Context, location string) (bool, error)

	// List returns all keys under prefix.
	List(ctx context.Context, prefix string) ([]string, error)
}
