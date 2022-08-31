package httputil

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
)

// FileServer is a specialized version of http.FileServer
// that assumes files rooted at FileSystem are name-hashed.
// cache-control are written specifically for index.html and name-hashed files.
type FileServer struct {
	FileSystem          http.FileSystem
	FallbackToIndexHTML bool
}

func (s *FileServer) writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, fs.ErrNotExist) {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, fs.ErrPermission) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	return
}

func (s *FileServer) open(name string) (http.File, fs.FileInfo, error) {
	file, err := s.FileSystem.Open(name)
	if err != nil {
		return nil, nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	// Unlike http.FileServer, we do not serve directory.
	if stat.IsDir() {
		file.Close()
		return nil, nil, fs.ErrNotExist
	}
	return file, stat, nil
}

func (s *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	indexHTML := "/index.html"

	// The treatment of path is different from http.FileServer.
	// We always normalize the path before we pass it to FileSystem.
	r.URL.Path = path.Clean("/" + r.URL.Path)

	// Rewrite path to index.html to handle HTML5 history routing convention.
	// err is used to determine whether we can cache-control header.
	file, stat, err := s.open(r.URL.Path)
	if s.FallbackToIndexHTML && errors.Is(err, fs.ErrNotExist) {
		r.URL.Path = indexHTML
		file, stat, err = s.open(r.URL.Path)
	}

	if err != nil {
		s.writeError(w, err)
		return
	}
	defer file.Close()

	// We only write cache-control header only when there is no error.
	isIndexHTML := r.URL.Path == indexHTML
	if isIndexHTML {
		// Force the browser to validate index.html
		w.Header().Set("Cache-Control", "no-cache")
	} else {
		// 7 Days
		w.Header().Set("Cache-Control", "public, max-age=604800")
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}
