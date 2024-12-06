package webapp

import (
	"net/http"
)

type ImplementationSwitcherHandler struct {
	Interaction http.Handler
	Authflow    http.Handler
	AuthflowV2  http.Handler
}

func (h *ImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.AuthflowV2.ServeHTTP(w, r)
}
