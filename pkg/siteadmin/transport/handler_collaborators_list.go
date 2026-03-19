package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorsListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/projects/:projectID/collaborators")
}

type CollaboratorsListHandler struct {
	// Add service dependencies here as needed
}

type CollaboratorsListParams struct {
	ProjectID string
}

func parseCollaboratorsListParams(r *http.Request) CollaboratorsListParams {
	return CollaboratorsListParams{
		ProjectID: httproute.GetParam(r, "projectID"),
	}
}

func (h *CollaboratorsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = parseCollaboratorsListParams(r)

	// TODO: implement
	http.NotFound(w, r)
}
