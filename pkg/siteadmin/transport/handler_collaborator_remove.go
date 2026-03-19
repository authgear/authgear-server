package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorRemoveRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("DELETE").
		WithPathPattern("/api/v1/projects/:projectID/collaborators/:collaboratorID")
}

type CollaboratorRemoveHandler struct {
	// Add service dependencies here as needed
}

type CollaboratorRemoveParams struct {
	ProjectID      string
	CollaboratorID string
}

func parseCollaboratorRemoveParams(r *http.Request) CollaboratorRemoveParams {
	return CollaboratorRemoveParams{
		ProjectID:      httproute.GetParam(r, "projectID"),
		CollaboratorID: httproute.GetParam(r, "collaboratorID"),
	}
}

func (h *CollaboratorRemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = parseCollaboratorRemoveParams(r)

	// TODO: implement
	http.NotFound(w, r)
}
