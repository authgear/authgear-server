package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureRuntimeConfigRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").WithPathPattern("/api/runtime-config.json")
}

type RuntimeConfig struct {
	AuthgearClientID string `json:"authgear_client_id"`
	AuthgearEndpoint string `json:"authgear_endpoint"`
}

type RuntimeConfigHandler struct {
	Config *config.ServerConfig
}

func (h *RuntimeConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := RuntimeConfig{
		AuthgearClientID: h.Config.Authgear.ClientID,
		AuthgearEndpoint: h.Config.Authgear.Endpoint,
	}

	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.Write(b)
}
