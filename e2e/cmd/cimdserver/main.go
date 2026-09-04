// cimdserver is a plain-HTTP document host for e2e CIMD tests. It runs on
// the host (like e2e-smtp/e2e-proxy), not in docker compose, because
// authgear itself -- which also runs on the host in e2e -- must be able to
// reach it as an unauthenticated fetch target the way a real CIMD document
// host would be reached. Plain HTTP, not HTTPS: the e2e projects that
// exercise it set insecure_http_allowed/insecure_fetch_address_allowed
// (docs/specs/cimd.md § SSRF Protection), so no TLS/CA handling is needed.
//
// Each test file configures its own document paths via the /_control/*
// endpoints before using them, and should use path names unique to that
// test file: e2e test files run with -parallel 5, and this server's state
// is process-wide, not per-test.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/authgear/authgear-server/pkg/util/debug"
)

const addr = "127.0.0.1:2727"

type pathConfig struct {
	// Status is the HTTP status to respond with. Zero means 200.
	Status int
	// ContentType overrides the response Content-Type. Empty means
	// application/json, the CIMD document default.
	ContentType string
	// Body is the raw response body to serve, used when Bytes and Redirect
	// are both zero/empty.
	Body []byte
	// Bytes, if non-zero, serves that many filler bytes instead of Body --
	// for the oversize-response case, and to prove size is enforced without
	// relying on the server ever sending a Content-Length header (it never
	// does; responses here are always chunked).
	Bytes int
	// Redirect, if non-empty, responds with a 301 to this path instead of
	// serving Body/Bytes.
	Redirect string
	// Offline, if true, blocks the request past any reasonable client
	// timeout (FetchTimeout is 5s), simulating a document host that never
	// responds.
	Offline bool
}

type server struct {
	mu      sync.Mutex
	configs map[string]*pathConfig
	hits    map[string]int
}

func newServer() *server {
	return &server{
		configs: map[string]*pathConfig{},
		hits:    map[string]int{},
	}
}

func (s *server) set(path string, cfg *pathConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[path] = cfg
}

func (s *server) hit(path string) *pathConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hits[path]++
	return s.configs[path]
}

func (s *server) hitCount(path string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hits[path]
}

func (s *server) reset(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.configs, path)
	delete(s.hits, path)
}

type setDocumentRequest struct {
	Path     string          `json:"path"`
	Status   int             `json:"status"`
	Document json.RawMessage `json:"document"`
	Body     string          `json:"body"`
	// BodyBase64, if non-empty, is base64-decoded and used as the response
	// body instead of Body -- for content that is not valid UTF-8 (e.g. a
	// real PNG's magic bytes), which cannot round-trip through a JSON
	// string.
	BodyBase64  string `json:"body_base64"`
	ContentType string `json:"content_type"`
}

type setBytesRequest struct {
	Path  string `json:"path"`
	Bytes int    `json:"bytes"`
}

type setRedirectRequest struct {
	Path string `json:"path"`
	To   string `json:"to"`
}

type pathRequest struct {
	Path string `json:"path"`
}

func writeJSONError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func decodeBody[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err)
		return req, false
	}
	return req, true
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleSet configures the document (or an arbitrary status) a path
// serves. Send either "document" (marshalled as the body, Content-Type:
// application/json) or a raw "body" string.
func (s *server) handleSet(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[setDocumentRequest](w, r)
	if !ok {
		return
	}
	body := []byte(req.Body)
	if req.BodyBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.BodyBase64)
		if err != nil {
			writeJSONError(w, err)
			return
		}
		body = decoded
	}
	if len(req.Document) > 0 {
		body = req.Document
	}
	status := req.Status
	if status == 0 {
		status = http.StatusOK
	}
	s.set(req.Path, &pathConfig{Status: status, Body: body, ContentType: req.ContentType})
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleSetBytes(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[setBytesRequest](w, r)
	if !ok {
		return
	}
	s.set(req.Path, &pathConfig{Bytes: req.Bytes})
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleSetRedirect(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[setRedirectRequest](w, r)
	if !ok {
		return
	}
	s.set(req.Path, &pathConfig{Redirect: req.To})
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleSetOffline(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[pathRequest](w, r)
	if !ok {
		return
	}
	s.set(req.Path, &pathConfig{Offline: true})
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleReset(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[pathRequest](w, r)
	if !ok {
		return
	}
	s.reset(req.Path)
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleHits(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Query().Get("path"), "/")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"hits": s.hitCount(path)})
}

func (s *server) handleDocument(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	cfg := s.hit(path)
	if cfg == nil {
		http.NotFound(w, r)
		return
	}

	switch {
	case cfg.Offline:
		select {
		case <-time.After(30 * time.Second):
		case <-r.Context().Done():
		}
	case cfg.Redirect != "":
		http.Redirect(w, r, cfg.Redirect, http.StatusMovedPermanently)
	default:
		writeDocument(w, cfg)
	}
}

// writeDocument serves cfg.Body, or cfg.Bytes filler bytes when set. Filler
// is written without a Content-Length header, so this exercises the same
// unknown-length/chunked path a real oversized response would take -- the
// client's size limit must not depend on a declared header.
func writeDocument(w http.ResponseWriter, cfg *pathConfig) {
	status := cfg.Status
	if status == 0 {
		status = http.StatusOK
	}
	contentType := cfg.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)

	if cfg.Bytes == 0 {
		_, _ = w.Write(cfg.Body)
		return
	}

	filler := bytes.Repeat([]byte{'a'}, 4096)
	remaining := cfg.Bytes
	for remaining > 0 {
		n := min(remaining, len(filler))
		if _, err := w.Write(filler[:n]); err != nil {
			return
		}
		remaining -= n
	}
}

func newMux(s *server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/_control/set", s.handleSet)
	mux.HandleFunc("/_control/set_bytes", s.handleSetBytes)
	mux.HandleFunc("/_control/set_redirect", s.handleSetRedirect)
	mux.HandleFunc("/_control/set_offline", s.handleSetOffline)
	mux.HandleFunc("/_control/reset", s.handleReset)
	mux.HandleFunc("/_control/hits", s.handleHits)
	mux.HandleFunc("/", s.handleDocument)
	return mux
}

func main() {
	debug.TrapSIGQUIT()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           newMux(newServer()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = httpServer.Shutdown(context.Background())
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start cimd document server: %v", err)
	}
}
