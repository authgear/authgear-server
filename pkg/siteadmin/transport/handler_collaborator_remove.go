package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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
	params := parseCollaboratorRemoveParams(r)

	// TODO: Replace with real data source. Search dummy data for now.
	for _, c := range dummyCollaboratorsForProject(params.ProjectID) {
		if c.Id == params.CollaboratorID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(struct{}{})
			return
		}
	}

	writeError(w, r, apierrors.NewNotFound("collaborator not found"))
}
