package webapp

import (
	"net/http"

	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/oauth2/callback/:alias")
}

type SSOCallbackInteractions interface {
	HandleOAuthCallback(state *interactionflows.State, data interactionflows.OAuthCallbackData, opts interactionflows.HandleOAuthCallbackOptions) (*interactionflows.WebAppResult, error)
}

type SSOCallbackHandler struct {
	Database     *db.Handle
	State        StateService
	Interactions SSOCallbackInteractions
	Responder    Responder
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providerAlias := httproute.GetParam(r, "alias")

	data := interactionflows.OAuthCallbackData{
		State:            r.Form.Get("state"),
		Code:             r.Form.Get("code"),
		Scope:            r.Form.Get("scope"),
		Error:            r.Form.Get("error"),
		ErrorDescription: r.Form.Get("error_description"),
	}

	// Add x_sid so CloneState works.
	q := r.URL.Query()
	q.Set("x_sid", data.State)
	u := *r.URL
	u.RawQuery = q.Encode()
	r.URL = &u

	h.Database.WithTx(func() error {
		state, err := h.State.CloneState(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		var result *interactionflows.WebAppResult
		defer func() {
			h.State.UpdateState(state, result, err)
			h.Responder.Respond(w, r, state, result, err)
		}()

		nonceSource, _ := r.Cookie(webapp.CSRFCookieName)
		result, err = h.Interactions.HandleOAuthCallback(state, data, interactionflows.HandleOAuthCallbackOptions{
			ProviderAlias: providerAlias,
			NonceSource:   nonceSource,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
