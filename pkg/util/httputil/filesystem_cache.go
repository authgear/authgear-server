package httputil

import (
	// nolint:gosec
	"crypto/md5"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/authgear/authgear-server/pkg/util/filepathutil"
)

// FilesystemCache is a helper to write the response into the tmp directory.
// The response is then served with http.FileServer,
// with the advantage of supporting range request and cache validation.
// If the file is not modified, the response is a 304.
// For even better performance, we need to add Cache-Control header
// to take advantage of the fact that the filename is hashed.
// However, http.FileServer does not support Cache-Control.
// Unconditionally adding Cache-Control for non-existent file is problematic.
type FilesystemCache struct {
	mutexForMapping sync.RWMutex
	mapping         map[string]string
	mutexForFile    sync.Mutex
}

func NewFilesystemCache() *FilesystemCache {
	return &FilesystemCache{
		mapping: make(map[string]string),
	}
}

func (c *FilesystemCache) makeFilePath(filename string) string {
	return filepath.Join(os.TempDir(), filename)
}

func (c *FilesystemCache) write(filePath string, bytes []byte) error {
	c.mutexForFile.Lock()
	defer c.mutexForFile.Unlock()
	// nolint: gosec
	return os.WriteFile(filePath, bytes, 0666)
}

func (c *FilesystemCache) Clear() error {
	c.mutexForFile.Lock()
	c.mutexForMapping.Lock()
	defer c.mutexForFile.Unlock()
	defer c.mutexForMapping.Unlock()

	for _, mappedFilename := range c.mapping {
		filePath := c.makeFilePath(mappedFilename)
		err := os.Remove(filePath)
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
		if err != nil {
			return err
		}
	}
	c.mapping = make(map[string]string)
	return nil
}

func (c *FilesystemCache) Serve(r *http.Request, make func() ([]byte, error)) (handler http.Handler) {
	var err error
	var bytes []byte

	originalPath := r.URL.Path
	filename := filepathutil.EscapePath(originalPath)

	c.mutexForMapping.RLock()
	mappedFilename, ok := c.mapping[filename]
	c.mutexForMapping.RUnlock()

	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok {
			r.URL.Path = fmt.Sprintf("/%v", mappedFilename)
		} else {
			// This will result in 404.
			r.URL.Path = fmt.Sprintf("/%v", filename)
		}

		http.FileServer(http.Dir(os.TempDir())).ServeHTTP(w, r)
	})

	needWrite := false
	if ok {
		filePath := c.makeFilePath(mappedFilename)
		_, err = os.Stat(filePath)
		if err != nil {
			needWrite = true
		}
	} else {
		needWrite = true
	}

	if needWrite {
		bytes, err = make()
		if err != nil {
			return
		}

		// nolint:gosec
		hashBytes := md5.Sum(bytes)
		hash := fmt.Sprintf("%x", hashBytes)
		mappedFilename := filepathutil.MakeHashedPath(filename, hash)
		filePath := c.makeFilePath(mappedFilename)
		err = c.write(filePath, bytes)
		if err != nil {
			return
		}

		c.mutexForMapping.Lock()
		c.mapping[filename] = mappedFilename
		c.mutexForMapping.Unlock()
	}

	return
}
