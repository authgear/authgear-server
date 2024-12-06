package webapp

import (
	"net/http"
)

type ImplementationSwitcherHandler struct {
	AuthflowV2 http.Handler
}

func (h *ImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.AuthflowV2.ServeHTTP(w, r)
}
