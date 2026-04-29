package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorPromoteRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/v1/apps/:appID/collaborators/:collaboratorID/promote")
}

type CollaboratorPromoteService interface {
	PromoteCollaborator(ctx context.Context, appID string, collaboratorID string) (*siteadmin.Collaborator, error)
}

type CollaboratorPromoteHandler struct {
	Service CollaboratorPromoteService
}

type CollaboratorPromoteParams struct {
	AppID          string
	CollaboratorID string
}

func parseCollaboratorPromoteParams(r *http.Request) CollaboratorPromoteParams {
	return CollaboratorPromoteParams{
		AppID:          httproute.GetParam(r, "appID"),
		CollaboratorID: httproute.GetParam(r, "collaboratorID"),
	}
}

func (h *CollaboratorPromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseCollaboratorPromoteParams(r)

	collaborator, err := h.Service.PromoteCollaborator(r.Context(), params.AppID, params.CollaboratorID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	SiteAdminAPISuccessResponse{Body: collaborator}.WriteTo(w)
}
