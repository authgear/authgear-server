package transport

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAppGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/apps/:appID")
}

type AppGetHandler struct {
	// Add service dependencies here as needed
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

	// TODO: Replace with real data source. Search dummy data for now.
	for _, a := range dummyApps {
		if strings.EqualFold(a.Id, params.AppID) {
			detail := siteadmin.AppDetail{
				Id:         a.Id,
				OwnerEmail: a.OwnerEmail,
				Plan:       a.Plan,
				CreatedAt:  a.CreatedAt,
				UserCount:  300,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(detail)
			return
		}
	}

	writeError(w, r, apierrors.NewNotFound("app not found"))
}
