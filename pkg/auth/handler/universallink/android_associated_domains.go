package universallink

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAndroidAssociatedDomainsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "OPTIONS").
		WithPathPattern("/.well-known/assetlinks.json")
}

type AndroidAssociatedDomainsProvider interface {
	PopulateAndroidAssociatedDomains(data *[]interface{})
}

type AndroidAssociatedDomainsHandler struct {
	Provider AndroidAssociatedDomainsProvider
}

func (h *AndroidAssociatedDomainsHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	data := []interface{}{}
	h.Provider.PopulateAndroidAssociatedDomains(&data)

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err := encoder.Encode(data)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}
