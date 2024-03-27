package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/oauth2/callback/:alias")
}

type SSOCallbackHandler struct {
	AuthflowController *AuthflowController
	ControllerFactory  ControllerFactory
	UIConfig           *config.UIConfig
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	state := r.Form.Get("state")
	code := r.Form.Get("code")
	error_ := r.Form.Get("error")
	errorDescription := r.Form.Get("error_description")
	errorURI := r.Form.Get("error_uri")

	switch h.UIConfig.Implementation.WithDefault() {
	case config.UIImplementationAuthflow:
		fallthrough
	case config.UIImplementationAuthflowV2:
		// authflow
		h.AuthflowController.HandleOAuthCallback(w, r, AuthflowOAuthCallbackResponse{
			State:            state,
			Code:             code,
			Error:            error_,
			ErrorDescription: errorDescription,
			ErrorURI:         errorURI,
		})
	default:
		// interaction
		ctrl, err := h.ControllerFactory.New(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer ctrl.Serve()

		data := InputOAuthCallback{
			ProviderAlias: httproute.GetParam(r, "alias"),

			Code:             code,
			Error:            error_,
			ErrorDescription: errorDescription,
			ErrorURI:         errorURI,
		}

		handler := func() error {
			result, err := ctrl.InteractionOAuthCallback(data, state)
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		}
		ctrl.Get(handler)
		ctrl.PostAction("", handler)
	}
}
