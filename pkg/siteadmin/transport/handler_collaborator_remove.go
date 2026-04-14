package transport

import (
	"context"
	"net/http"

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

	if err := h.Service.RemoveCollaborator(r.Context(), params.AppID, params.CollaboratorID); err != nil {
		writeError(w, r, err)
		return
	}

	SiteAdminAPISuccessResponse{Body: struct{}{}}.WriteTo(w)
}
