package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureRegisterRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/oauth2/register")
}

type ProtocolRegistrationHandler interface {
	Handle(ctx context.Context, r *http.Request) (*handler.RegistrationResponse, error)
}

type RegisterHandler struct {
	RegistrationHandler ProtocolRegistrationHandler
}

func (h *RegisterHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// #nosec G120 -- BodyLimitMiddleware caps POST bodies to 1MB for this
	// endpoint, applied at the base middleware chain.
	resp, err := h.RegistrationHandler.Handle(r.Context(), r)
	if err != nil {
		var oauthErr *protocol.OAuthProtocolError
		if errors.As(err, &oauthErr) {
			status := oauthErr.StatusCode
			if status == 0 {
				status = http.StatusBadRequest
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(status)
			_ = json.NewEncoder(rw).Encode(oauthErr.Response)
			return
		}
		http.Error(rw, "Internal Server Error", 500)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(rw).Encode(resp)
}
