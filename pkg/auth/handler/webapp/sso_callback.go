package webapp

import (
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/oauth2/callback/:alias")
}

type SSOCallbackHandler struct {
	ControllerFactory ControllerFactory
	CSRFCookie        webapp.CSRFCookieDef
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nonceSource, _ := r.Cookie(h.CSRFCookie.Name)

	data := InputOAuthCallback{
		ProviderAlias: httproute.GetParam(r, "alias"),
		NonceSource:   nonceSource,

		Code:             r.Form.Get("code"),
		Scope:            r.Form.Get("scope"),
		Error:            r.Form.Get("error"),
		ErrorDescription: r.Form.Get("error_description"),
	}

	handler := func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		// Always use the current step of session
		r.Form.Set("x_step", strconv.Itoa(len(session.Steps)))

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
