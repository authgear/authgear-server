package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureSystemConfigRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").WithPathPattern("/api/system-config.json")
}

type SystemConfigProvider interface {
	SystemConfig(ctx context.Context) (*model.SystemConfig, error)
}

type SystemConfigHandler struct {
	SystemConfig    SystemConfigProvider
	FilesystemCache *httputil.FilesystemCache
}

func (h *SystemConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	make := func() ([]byte, error) {
		ctx := r.Context()
		cfg, err := h.SystemConfig.SystemConfig(ctx)
		if err != nil {
			return nil, err
		}

		b, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}

		return b, nil
	}

	h.FilesystemCache.Serve(r, make).ServeHTTP(w, r)
}
