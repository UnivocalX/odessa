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
	root, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(location, "file://") {
		u, err := url.Parse(location)
		if err != nil {
			return "", err
		}
		if u.Host != "" && u.Host != "localhost" {
			return "", os.ErrPermission
		}
		location = u.Path
	}

	location = filepath.Clean(filepath.FromSlash(location))
	location = strings.TrimLeft(location, string(filepath.Separator))
	path := filepath.Join(root, location)
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", os.ErrPermission
	}

	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	probe := path
	for {
		resolvedPath, resolveErr := filepath.EvalSymlinks(probe)
		if resolveErr == nil {
			rel, relErr := filepath.Rel(resolvedRoot, resolvedPath)
			if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return "", os.ErrPermission
			}
			break
		}
		if !os.IsNotExist(resolveErr) {
			return "", resolveErr
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			break
		}
		probe = parent
	}

	return path, nil
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
