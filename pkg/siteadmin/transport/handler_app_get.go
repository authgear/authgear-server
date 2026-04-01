package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAppGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/apps/:appID")
}

type AppGetService interface {
	GetApp(ctx context.Context, appID string) (*siteadmin.AppDetail, error)
}

type AppGetHandler struct {
	AppGet AppGetService
}

type AppGetParams struct {
	AppID string
}

func parseAppGetParams(r *http.Request) AppGetParams {
	return AppGetParams{
		AppID: httproute.GetParam(r, "appID"),
	}
}

func (h *AppGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseAppGetParams(r)

	detail, err := h.AppGet.GetApp(r.Context(), params.AppID)
	if err != nil {
		if errors.Is(err, configsource.ErrAppNotFound) {
			writeError(w, r, apierrors.NewNotFound("app not found"))
			return
		}
		writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(detail)
}
