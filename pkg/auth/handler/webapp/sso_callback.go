package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/oauth2/callback/:alias")
}

type SSOProvider interface {
	HandleSSOCallback(w http.ResponseWriter, r *http.Request, providerAlias string) (func(error), error)
}

type SSOCallbackHandler struct {
	Provider SSOProvider
	Database *db.Handle
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providerAlias := httproute.GetParam(r, "alias")

	h.Database.WithTx(func() error {
		writeResponse, err := h.Provider.HandleSSOCallback(w, r, providerAlias)
		writeResponse(err)
		return err
	})
}
