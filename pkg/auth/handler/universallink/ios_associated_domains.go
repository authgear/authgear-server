package universallink

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureIOSAssociatedDomainsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "OPTIONS").
		WithPathPattern("/.well-known/apple-app-site-association")
}

type IOSAssociatedDomainsProvider interface {
	PopulateIOSAssociatedDomains(data map[string]interface{})
}

type IOSAssociatedDomainsHandler struct {
	Provider IOSAssociatedDomainsProvider
}

func (h *IOSAssociatedDomainsHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	h.Provider.PopulateIOSAssociatedDomains(data)

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err := encoder.Encode(data)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}
