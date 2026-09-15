package httputil

import (
	"bytes"
	"crypto/subtle"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/util/filepathutil"
)

const sourceMapBasicAuthRealm = "source map"

// SourceMapConfig configures how FileServer serves source map (*.map) files.
type SourceMapConfig struct {
	// Enabled sets whether source map files are served at all.
	// When it is false, source map files are served as if they did not exist.
	Enabled bool
	// SentryToken, when non-empty, protects source map files with HTTP basic authentication.
	// The token is the password, and the username is ignored.
	// Sentry fetches a publicly hosted source map with a configurable header,
	// so it can be configured to send `Authorization: Basic <base64 of ":<token>">`.
	// See https://docs.sentry.io/platforms/javascript/sourcemaps/uploading/hosting-publicly/
	SentryToken string
}

// IsAuthorized tells whether r is allowed to fetch a source map file.
func (c SourceMapConfig) IsAuthorized(r *http.Request) bool {
	if c.SentryToken == "" {
		return true
	}

	_, password, ok := r.BasicAuth()
	if !ok {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(password), []byte(c.SentryToken)) == 1
}

type FileServerIndexHTMLTemplateDataKeyType struct{}

var FileServerIndexHTMLtemplateDataKey = FileServerIndexHTMLTemplateDataKeyType{}

// FileServer is a specialized version of http.FileServer
// that assumes files rooted at FileSystem are name-hashed.
// Cache-control are written specifically for index.html and name-hashed files.
// When serving index.html, index.html is assumed to be a Go template.
// FileServer will use the context value FileServerIndexHTMLTemplateDataKey to render.
type FileServer struct {
	FileSystem          http.FileSystem
	AssetsDir           string
	FallbackToIndexHTML bool
	// SourceMap configures how source map (*.map) files are served.
	// The zero value serves no source map file.
	SourceMap SourceMapConfig
}

func (s *FileServer) writeError(w http.ResponseWriter, err error) {
	// http.Error is NOT used intentionally to avoid returning a text/plain response.
	// The desired response is WITHOUT content-type, and with content-length: 0
	w.Header().Del("Content-Type")
	w.Header().Set("Content-Length", "0")
	if errors.Is(err, fs.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, fs.ErrPermission) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	return
}

func (s *FileServer) writeUnauthorized(w http.ResponseWriter) {
	// http.Error is NOT used intentionally to avoid returning a text/plain response.
	// The desired response is WITHOUT content-type, and with content-length: 0
	w.Header().Del("Content-Type")
	w.Header().Set("Content-Length", "0")
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q", sourceMapBasicAuthRealm))
	w.WriteHeader(http.StatusUnauthorized)
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

func (s *FileServer) serveAsset(w http.ResponseWriter, r *http.Request, isSourceMap bool) {
	file, stat, err := s.open(r.URL.Path)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer file.Close()

	if isSourceMap {
		// A source map response can be protected by credentials,
		// so it must never be stored by a shared cache.
		w.Header().Set("Cache-Control", "no-store")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=604800")
	}
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func (s *FileServer) serveOther(w http.ResponseWriter, r *http.Request) {
	file, stat, err := s.open(r.URL.Path)
	if err != nil {
		// The error is not file not found. Just report the error.
		if !errors.Is(err, fs.ErrNotExist) {
			s.writeError(w, err)
			return
		}

		// Otherwise the file is not found.
		// Just report the error if fallback to index.html is disabled.
		if !s.FallbackToIndexHTML {
			s.writeError(w, err)
			return
		}

		r.URL.Path = "/index.html"
		indexHTMLFile, indexHTMLStat, err := s.open(r.URL.Path)
		// No idea how to handle, just report the error.
		if err != nil {
			s.writeError(w, err)
			return
		}
		defer indexHTMLFile.Close()
		// Serve index.html

		indexHTMLBytes, err := io.ReadAll(indexHTMLFile)
		if err != nil {
			// index.html exists but not readable.
			panic(err)
		}

		tpl, err := htmltemplate.New("").Parse(string(indexHTMLBytes))
		if err != nil {
			// We panic because this prints a stack trace to tell what is wrong with index.html.
			// This is more useful than just return 500.
			panic(err)
		}

		data := r.Context().Value(FileServerIndexHTMLtemplateDataKey)
		var buf bytes.Buffer
		err = tpl.Execute(&buf, data)
		if err != nil {
			panic(err)
		}

		// Use a zero modtime to ask http.ServeContent NOT to write Last-Modified.
		var modtime time.Time
		readSeeker := bytes.NewReader(buf.Bytes())
		http.ServeContent(w, r, indexHTMLStat.Name(), modtime, readSeeker)
		return
	}

	// Serve the original file.
	defer file.Close()
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func (s *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// By default, all response requires validation.
	// So a 404 response also requires validation.
	w.Header().Set("Cache-Control", "no-cache")

	// The treatment of path is different from http.FileServer.
	// We always normalize the path before we pass it to FileSystem.
	r.URL.Path = path.Clean("/" + r.URL.Path)

	// Source map files embed the application source code, so they are not served
	// unless they are explicitly enabled, and they can additionally be protected
	// by HTTP basic authentication.
	isSourceMap := filepathutil.IsSourceMapPath(r.URL.Path)
	if isSourceMap {
		if !s.SourceMap.Enabled {
			// Behave as if the source map file did not exist.
			s.writeError(w, fs.ErrNotExist)
			return
		}
		if !s.SourceMap.IsAuthorized(r) {
			s.writeUnauthorized(w)
			return
		}
	}

	// First of all we need to identity whether the path
	// seems like fetching a name-hashed file.
	//
	// If the request fetches a name-hashed file,
	// we return 404 for not found, 200 cache-control: public for found.
	//
	// If the request fetches a non-name-hashed file,
	// we fallback to index.html for not found.
	if strings.HasPrefix(r.URL.Path, "/"+s.AssetsDir) {
		s.serveAsset(w, r, isSourceMap)
	} else {
		s.serveOther(w, r)
	}
}
