// Package filestore keeps uploaded files on local disk. Callers only ever hold
// an opaque key returned by Save; the on-disk location never leaves this
// package, and every key is resolved back under the root directory.
package filestore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("file not found")

type Local struct {
	root string
}

// NewLocal creates the root directory if it doesn't exist yet.
func NewLocal(root string) (*Local, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, err
	}
	return &Local{root: abs}, nil
}

// Save writes r under a fresh random name inside folder and returns its key.
// The original file name is never used on disk, so it can't collide, traverse
// out of the root, or smuggle an executable extension; ext is taken from a
// caller-vetted allowlist.
func (s *Local) Save(folder, ext string, r io.Reader) (string, error) {
	key := filepath.ToSlash(filepath.Join(folder, uuid.NewString()+ext))

	path, err := s.resolve(key)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "", err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(file, r); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}

	return key, nil
}

func (s *Local) Read(key string) ([]byte, error) {
	path, err := s.resolve(key)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

// Delete is idempotent: removing a file that is already gone is not an error.
func (s *Local) Delete(key string) error {
	path, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// resolve maps a key to a path and refuses anything that would land outside
// the root.
func (s *Local) resolve(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("filestore: empty key")
	}

	path := filepath.Join(s.root, filepath.FromSlash(key))
	if path != s.root && !strings.HasPrefix(path, s.root+string(filepath.Separator)) {
		return "", fmt.Errorf("filestore: key %q escapes the storage root", key)
	}
	return path, nil
}
