package authflowv2

import (
	"net/http"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type AuthflowV2SignupHandler struct {
	SignupLoginHandler InternalAuthflowV2SignupLoginHandler
}

func (h *AuthflowV2SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.SignupLoginHandler.ServeHTTP(w, r, authflow.FlowTypeSignup)
}
