package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

type recorder struct {
	mu       sync.Mutex
	requests map[string][]map[string]interface{}
}

func newRecorder() *recorder {
	return &recorder{
		requests: map[string][]map[string]interface{}{},
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
			defer r.Body.Close()
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			rec.append(path, payload)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"ok"}`))
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"requests": rec.get(path),
			})
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
