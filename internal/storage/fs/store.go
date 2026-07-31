package fs

import (
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func (s *Store) abs(location string) (string, error) {
	if strings.HasPrefix(location, "file://") {
		u, err := url.Parse(location)
		if err != nil {
			return "", err
		}
		location = u.Path
	}
	return filepath.Join(s.root, location), nil
}

func (s *Store) Get(ctx context.Context, location string) (io.ReadCloser, error) {
	path, err := s.abs(location)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *Store) Put(ctx context.Context, location string, r io.Reader) error {
	path, err := s.abs(location)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *Store) Delete(ctx context.Context, location string) error {
	path, err := s.abs(location)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (s *Store) Available(ctx context.Context, location string) (bool, error) {
	path, err := s.abs(location)
	if err != nil {
		return false, err
	}
	
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	base, err := s.abs(prefix)
	if err != nil {
		return nil, err
	}
	var keys []string
	err = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			rel, relErr := filepath.Rel(s.root, path)
			if relErr != nil {
				return relErr
			}
			keys = append(keys, rel)
		}
		return nil
	})
	return keys, err
}
