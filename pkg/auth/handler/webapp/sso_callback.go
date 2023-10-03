package webapp

import (
	"net/http"

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

	switch {
	case state != "":
		// authflow
		s, err := h.AuthflowController.GetWebSession(r)
		if err != nil {
			h.AuthflowController.RenderError(w, r, err)
			return
		}

		xStep := state
		screen, err := h.AuthflowController.GetScreen(s, xStep)
		if err != nil {
			h.AuthflowController.RenderError(w, r, err)
			return
		}

		input := map[string]interface{}{}
		switch {
		case code != "":
			input["code"] = code
		case error_ != "":
			input["error"] = error_
			input["error_description"] = errorDescription
			input["error_uri"] = errorURI
		}
		result, err := h.AuthflowController.FeedInput(r, s, screen, input)
		if err != nil {
			h.AuthflowController.RenderError(w, r, err)
			return
		}

		result.WriteResponse(w, r)
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
			result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
				input = &data
				return
			})
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
