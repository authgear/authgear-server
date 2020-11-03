package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/oauth2/callback/:alias")
}

type SSOCallbackHandler struct {
	Database   *db.Handle
	WebApp     WebAppService
	CSRFCookie webapp.CSRFCookieDef
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nonceSource, _ := r.Cookie(h.CSRFCookie.Name)

	stateID := r.Form.Get("state")
	data := InputOAuthCallback{
		ProviderAlias: httproute.GetParam(r, "alias"),
		NonceSource:   nonceSource,

		Code:             r.Form.Get("code"),
		Scope:            r.Form.Get("scope"),
		Error:            r.Form.Get("error"),
		ErrorDescription: r.Form.Get("error_description"),
	}

	err := h.Database.WithTx(func() error {
		result, err := h.WebApp.PostInput(stateID, func() (input interface{}, err error) {
			input = &data
			return
		})
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})
	if err != nil {
		panic(err)
	}
}
