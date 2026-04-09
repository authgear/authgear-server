package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorRemoveRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "DELETE").
		WithPathPattern("/api/v1/apps/:appID/collaborators/:collaboratorID")
}

type CollaboratorRemoveHandler struct {
	Service CollaboratorRemoveService
}

type CollaboratorRemoveService interface {
	RemoveCollaborator(ctx context.Context, appID string, collaboratorID string) error
}

type CollaboratorRemoveParams struct {
	AppID          string
	CollaboratorID string
}

func parseCollaboratorRemoveParams(r *http.Request) CollaboratorRemoveParams {
	return CollaboratorRemoveParams{
		AppID:          httproute.GetParam(r, "appID"),
		CollaboratorID: httproute.GetParam(r, "collaboratorID"),
	}
}

func (h *CollaboratorRemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseCollaboratorRemoveParams(r)

	for _, c := range dummyCollaboratorsForApp(params.AppID) {
		if c.Id == params.CollaboratorID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(struct{}{})
			return
		}
	}

	writeError(w, r, apierrors.NewNotFound("collaborator not found"))
}
