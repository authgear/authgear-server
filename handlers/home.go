package handlers

import (
	"net/http"
)

// HomeHandler temp landing. FIXME
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Developer"))
}
