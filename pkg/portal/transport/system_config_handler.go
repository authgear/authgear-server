package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureSystemConfigRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").WithPathPattern("/api/system-config.json")
}

type SystemConfigProvider interface {
	SystemConfig() (*model.SystemConfig, error)
}

type SystemConfigHandler struct {
	SystemConfig SystemConfigProvider
}

func (h *SystemConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.SystemConfig.SystemConfig()
	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	if _, err := w.Write(b); err != nil {
		panic(err)
	}
}
