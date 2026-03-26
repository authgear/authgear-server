package transport

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureProjectGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/projects/:projectID")
}

type ProjectGetHandler struct {
	// Add service dependencies here as needed
}

type ProjectGetParams struct {
	ProjectID string
}

func parseProjectGetParams(r *http.Request) ProjectGetParams {
	return ProjectGetParams{
		ProjectID: httproute.GetParam(r, "projectID"),
	}
}

func (h *ProjectGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseProjectGetParams(r)

	// TODO: Replace with real data source. Search dummy data for now.
	for _, p := range dummyProjects {
		if strings.EqualFold(p.Id, params.ProjectID) {
			detail := siteadmin.ProjectDetail{
				Id:         p.Id,
				OwnerEmail: p.OwnerEmail,
				Plan:       p.Plan,
				CreatedAt:  p.CreatedAt,
				UserCount:  300,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(detail)
			return
		}
	}

	writeError(w, r, apierrors.NewNotFound("project not found"))
}
