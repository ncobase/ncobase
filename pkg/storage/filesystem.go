package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/casdoor/oss"
)

// FileSystem represents the interface for file system storage.
type FileSystem interface {
	GetFullPath(p string) string
	Get(p string) (*os.File, error)
	GetStream(p string) (io.ReadCloser, error)
	Put(p string, r io.Reader) (*oss.Object, error)
	Delete(p string) error
	List(p string) ([]*oss.Object, error)
	GetEndpoint() string
	GetURL(p string) (string, error)
}

// LocalFileSystem implements the FileSystem interface for local file system storage.
type LocalFileSystem struct {
	Folder string
}

// NewFileSystem creates a new local file system storage.
// It ensures the folder exists; if not, it creates the folder.
func NewFileSystem(folder string) *LocalFileSystem {
	abs, err := filepath.Abs(folder)
	if err != nil {
		panic("local file system storage's base folder is not initialized")
	}
	return &LocalFileSystem{Folder: abs}
}

// GetFullPath returns the full path from absolute / relative path.
func (fs *LocalFileSystem) GetFullPath(p string) string {
	fp := p
	if !strings.HasPrefix(p, fs.Folder) {
		fp, _ = filepath.Abs(filepath.Join(fs.Folder, p))
	}
	return fp
}

// Get receives a file with the given path.
func (fs *LocalFileSystem) Get(p string) (*os.File, error) {
	return os.Open(fs.GetFullPath(p))
}

// GetStream gets a file as a stream.
func (fs *LocalFileSystem) GetStream(p string) (io.ReadCloser, error) {
	return os.Open(fs.GetFullPath(p))
}

// Put stores the reader into the given path.
func (fs *LocalFileSystem) Put(p string, r io.Reader) (*oss.Object, error) {
	fp := fs.GetFullPath(p)
	if err := os.MkdirAll(filepath.Dir(fp), os.ModePerm); err != nil {
		return nil, err
	}

	dst, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, r)
	if err != nil {
		return nil, err
	}

	return &oss.Object{Path: p, Name: filepath.Base(p), StorageInterface: fs}, nil
}

// Delete deletes a file.
func (fs *LocalFileSystem) Delete(p string) error {
	return os.Remove(fs.GetFullPath(p))
}

// List lists files.
func (fs *LocalFileSystem) List(p string) ([]*oss.Object, error) {
	var (
		objects []*oss.Object
		fp      = fs.GetFullPath(p)
	)

	_ = filepath.Walk(fp, func(p string, info os.FileInfo, err error) error {
		if p == fp {
			return nil
		}

		if err == nil && !info.IsDir() {
			mt := info.ModTime()
			objects = append(objects, &oss.Object{
				Path:             strings.TrimPrefix(p, fs.Folder),
				Name:             info.Name(),
				LastModified:     &mt,
				StorageInterface: fs,
			})
		}
		return nil
	})

	return objects, nil
}

// GetEndpoint gets the endpoint. For FileSystem, the endpoint is "/".
func (fs *LocalFileSystem) GetEndpoint() string {
	return "/"
}

// GetURL gets the public accessible URL.
func (fs *LocalFileSystem) GetURL(p string) (string, error) {
	return p, nil
}
