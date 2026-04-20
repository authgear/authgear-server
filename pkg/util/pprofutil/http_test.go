package pprofutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServeMuxMetrics(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	NewServeMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "go_memstats_heap_inuse_bytes") {
		t.Fatalf("expected go memstats metrics in response")
	}
	if !strings.Contains(body, "go_threads") {
		t.Fatalf("expected go runtime metrics in response")
	}
}
