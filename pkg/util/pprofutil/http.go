package pprofutil

import (
	"net/http"
	"net/http/pprof"
)

// NewServeMux creates a new ServeMux and replicates the effect of net/http/pprof.init().
// See https://cs.opensource.google/go/go/+/refs/tags/go1.21.4:src/net/http/pprof/pprof.go;l=94
func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}
