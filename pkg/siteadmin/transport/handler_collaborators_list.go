package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/siteadmin/model"
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
	params := parseCollaboratorsListParams(r)

	// TODO: Replace with real data source. Search dummy data for now.
	for _, p := range dummyProjects {
		if p.Id == params.ProjectID {
			response := model.CollaboratorsListResponse{
				Collaborators: dummyCollaboratorsForProject(params.ProjectID),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
			return
		}
	}

	writeError(w, r, apierrors.NewNotFound("project not found"))
}
