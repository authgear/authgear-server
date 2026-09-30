package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// smsGatewayPathPrefix marks the paths that answer as an SMS gateway. No
// event-hook receiver uses it, so the two shapes never meet.
const smsGatewayPathPrefix = "sms-gateway/"

// logsPathSuffix marks the paths that answer as a Datadog logs intake: a
// path ending in /logs is answered 202, the status the sender treats as
// success, following the same path-conditional shape smsGatewayPathPrefix
// already uses for its own response.
const logsPathSuffix = "/logs"

// minCountPollInterval and minCountPollTimeout bound how long GET
// /<path>?min_count=<n> waits for records to arrive. Streaming delivery is
// asynchronous behind the drain interval, unlike a hook's synchronous
// delivery, so a caller needs to wait rather than poll from the test
// runner side.
const (
	minCountPollInterval = 50 * time.Millisecond
	minCountPollTimeout  = 10 * time.Second
)

type recorder struct {
	mu       sync.Mutex
	requests map[string][]map[string]interface{}
	// requestCount is a per-path counter of POST requests handled, used
	// as __request.index -- every record exploded from one request's
	// array body shares the same index, so a test can tell whether two
	// records arrived together or across separate requests.
	requestCount map[string]int
}

func newRecorder() *recorder {
	return &recorder{
		requests:     map[string][]map[string]interface{}{},
		requestCount: map[string]int{},
	}
}

func (r *recorder) append(key string, payload map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[key] = append(r.requests[key], payload)
}

func (r *recorder) get(key string) []map[string]interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	requests := r.requests[key]
	out := make([]map[string]interface{}, len(requests))
	copy(out, requests)
	return out
}

// nextRequestIndex returns the 0-indexed sequence number of the next POST
// request to key, and records that one more has now been seen.
func (r *recorder) nextRequestIndex(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := r.requestCount[key]
	r.requestCount[key] = n + 1
	return n
}

func trapSIGQUIT() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT)
	go func() {
		for range c {
			buf := make([]byte, 1024)
			for {
				n := runtime.Stack(buf, true)
				if n < len(buf) {
					_, _ = os.Stderr.Write(buf[:n])
					break
				}
				buf = make([]byte, 2*len(buf))
			}
		}
	}()
}

// decodeBody gunzips body when contentEncoding is gzip, then decodes it as
// either a single JSON object or an array of objects -- a Datadog intake
// request body is a gzipped JSON array, everything else here is a plain
// JSON object, and both end up as a slice of records so the caller does
// not need to know which shape it received.
func decodeBody(body io.Reader, contentEncoding string) ([]map[string]interface{}, error) {
	if contentEncoding == "gzip" {
		gz, err := gzip.NewReader(body)
		if err != nil {
			return nil, err
		}
		defer func() { _ = gz.Close() }()
		body = gz
	}

	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var decoded interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}

	switch v := decoded.(type) {
	case []interface{}:
		records := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("array element is not a JSON object")
			}
			records = append(records, m)
		}
		return records, nil
	case map[string]interface{}:
		return []map[string]interface{}{v}, nil
	default:
		return nil, fmt.Errorf("body must be a JSON object or array")
	}
}

// requestHeaders returns h as a flat map of header name to its first
// value, for embedding in the __request envelope. Header names are
// already canonicalised by net/http.
func requestHeaders(h http.Header) map[string]interface{} {
	out := make(map[string]interface{}, len(h))
	for name := range h {
		out[name] = h.Get(name)
	}
	return out
}

func main() {
	trapSIGQUIT()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rec := newRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodPost:
			handlePost(rec, w, r, path)
		case http.MethodGet:
			handleGet(rec, w, r, path)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{
		Addr:    "0.0.0.0:2626",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start hook server: %v", err)
	}
}

func handlePost(rec *recorder, w http.ResponseWriter, r *http.Request, path string) {
	defer r.Body.Close()

	contentEncoding := r.Header.Get("Content-Encoding")
	payloads, err := decodeBody(r.Body, contentEncoding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	envelope := map[string]interface{}{
		"index":            rec.nextRequestIndex(path),
		"method":           r.Method,
		"content_type":     r.Header.Get("Content-Type"),
		"content_encoding": contentEncoding,
		"headers":          requestHeaders(r.Header),
	}
	for _, payload := range payloads {
		payload["__request"] = envelope
		rec.append(path, payload)
	}

	switch {
	case strings.HasSuffix(path, logsPathSuffix):
		// The real Datadog intake answers 202; the sender treats anything
		// else as a failed chunk, so answering 200 here would make a test
		// pass while production behaviour was broken.
		w.WriteHeader(http.StatusAccepted)
	case strings.HasPrefix(path, smsGatewayPathPrefix):
		// A custom SMS provider reads the response body and treats
		// anything but code "ok" as a send failure, so a path under
		// sms-gateway/ answers in the shape docs/specs/sms_gateway.md
		// defines rather than the generic one.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"ok"}`))
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	}
}

func handleGet(rec *recorder, w http.ResponseWriter, r *http.Request, path string) {
	minCount, _ := strconv.Atoi(r.URL.Query().Get("min_count"))
	if minCount > 0 {
		deadline := time.Now().Add(minCountPollTimeout)
		for len(rec.get(path)) < minCount && time.Now().Before(deadline) {
			time.Sleep(minCountPollInterval)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": rec.get(path),
	})
}
