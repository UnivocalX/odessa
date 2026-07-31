package fs

import (
	"fmt"

	"github.com/UnivocalX/odessa/internal/storage"
)

func init() {
	storage.Register("file", &Store{root: "~/"})
}

type Store struct {
	root string
}

type Option func(*Store) error

func WithRoot(path string) Option {
	return func(s *Store) error {
		s.root = path
		return nil
	}
}

// Configure applies options to the globally-registered filesystem store.
func Configure(opts ...Option) error {
	backend, ok := storage.Backend("file")
	if !ok {
		return fmt.Errorf("fs: backend not registered")
	}
	s := backend.(*Store)
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return err
		}
	}
	return nil
}
