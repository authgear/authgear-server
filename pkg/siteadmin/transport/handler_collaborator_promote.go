package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorPromoteRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/v1/apps/:appID/collaborators/:collaboratorID/promote")
}

type CollaboratorPromoteHandler struct {
	// Service dependencies added in Stage 5
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
	_ = parseCollaboratorPromoteParams(r)
	http.NotFound(w, r)
}
