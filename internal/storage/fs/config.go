package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnivocalX/odessa/internal/storage"
)

func init() {
	storage.Register("file", &Store{root: "."})
}

type Store struct {
	root string
}

type Option func(*Store) error

func WithRoot(path string) Option {
	return func(s *Store) error {
		if strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
		root, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		root = filepath.Clean(root)
		info, err := os.Stat(root)
		if err != nil {
			return fmt.Errorf("fs: stat root: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("fs: root is not a directory")
		}
		resolvedRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			return fmt.Errorf("fs: resolve root: %w", err)
		}
		s.root = filepath.Clean(resolvedRoot)
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
