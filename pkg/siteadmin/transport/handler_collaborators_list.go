package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorsListRoute(route httproute.Route) httproute.Route {
	// The OPTIONS request is handled in CollaboratorAddRoute
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/apps/:appID/collaborators")
}

type CollaboratorsListHandler struct {
	// Add service dependencies here as needed
}

type CollaboratorsListParams struct {
	AppID string
}

func parseCollaboratorsListParams(r *http.Request) CollaboratorsListParams {
	return CollaboratorsListParams{
		AppID: httproute.GetParam(r, "appID"),
	}
}

func (h *CollaboratorsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseCollaboratorsListParams(r)

	// TODO: Replace with real data source. Search dummy data for now.
	for _, a := range dummyApps {
		if a.Id == params.AppID {
			response := siteadmin.CollaboratorsListResponse{
				Collaborators: dummyCollaboratorsForApp(params.AppID),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
			return
		}
	}

	writeError(w, r, apierrors.NewNotFound("app not found"))
}
