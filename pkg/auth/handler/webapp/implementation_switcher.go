package webapp

import (
	"net/http"
)

type ImplementationSwitcherMiddleware struct{}

func (m *ImplementationSwitcherMiddleware) Handle(next http.Handler) http.Handler {
	return next
}

type ImplementationSwitcherHandler struct {
	Interaction http.Handler
	Authflow    http.Handler
	AuthflowV2  http.Handler
}

func (h *ImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.AuthflowV2.ServeHTTP(w, r)
}
