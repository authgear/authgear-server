package httputil

import (
	"bytes"
	"errors"
	htmltemplate "html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"time"
)

type FileServerIndexHTMLTemplateDataKeyType struct{}

var FileServerIndexHTMLtemplateDataKey = FileServerIndexHTMLTemplateDataKeyType{}

var hashRegexp = regexp.MustCompile(`-[-_A-Za-z0-9]{8}`)

func Ext(basename string) string {
	stdlibPathExt := path.Ext(basename)
	// Handle cases like .vimrc
	// .vimrc is not considered as a file extension
	if stdlibPathExt == basename {
		return ""
	}

	return stdlibPathExt
}

// IsLikeRollupDefaultAssetName returns whether p looks like the default asset name
// documented in https://rollupjs.org/configuration-options/#output-assetfilenames
// That is, [name]-[hash][extname]
//
// Here are some examples
// appListQuery.js           -> appListQuery-nhRGQ3z4.js and appListQuery-nhRGQ3z4.js.map
// appListQuery.generated.js -> appListQuery.generated-DQp0G87L.js and appListQuery.generated-DQp0G87L.js.map
//
// The default length of hash is 8.
// See https://github.com/rollup/rollup/blob/v4.27.4/src/utils/hashPlaceholders.ts#L12
//
// After observing the examples, we know that splitting on dot does not always work, when the original file contains more than 1 dot.
// So a simpler heuristic is to check whether the filename has the pattern -[-_A-Za-z0-9]{8} and it has a file extension.
func IsLikeRollupDefaultAssetName(p string) bool {
	basename := path.Base(p)
	result := hashRegexp.FindStringSubmatch(basename)

	if len(result) == 0 {
		return false
	}

	ext := Ext(basename)
	if ext == "" {
		return false
	}

	return true
}

// FileServer is a specialized version of http.FileServer
// that assumes files rooted at FileSystem are name-hashed.
// Cache-control are written specifically for index.html and name-hashed files.
// When serving index.html, index.html is assumed to be a Go template.
// FileServer will use the context value FileServerIndexHTMLTemplateDataKey to render.
type FileServer struct {
	FileSystem          http.FileSystem
	FallbackToIndexHTML bool
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

func (s *FileServer) serveNameHashed(w http.ResponseWriter, r *http.Request) {
	file, stat, err := s.open(r.URL.Path)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer file.Close()

	w.Header().Set("Cache-Control", "public, max-age=604800")
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

	// First of all we need to identity whether the path
	// seems like fetching a name-hashed file.
	//
	// If the request fetches a name-hashed file,
	// we return 404 for not found, 200 cache-control: public for found.
	//
	// If the request fetches a non-name-hashed file,
	// we fallback to index.html for not found.
	if IsLikeRollupDefaultAssetName(r.URL.Path) {
		s.serveNameHashed(w, r)
	} else {
		s.serveOther(w, r)
	}
}
