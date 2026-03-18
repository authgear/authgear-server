package transport

import (
	"net/http"

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
	_ = parseProjectGetParams(r)

	// TODO: Replace with real data source.
	http.NotFound(w, r)
}
