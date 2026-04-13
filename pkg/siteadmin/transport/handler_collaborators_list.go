package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorsListRoute(route httproute.Route) httproute.Route {
	// The OPTIONS request is handled in CollaboratorAddRoute
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/apps/:appID/collaborators")
}

type CollaboratorsListHandler struct {
	Service CollaboratorsListService
}

type CollaboratorsListService interface {
	ListCollaborators(ctx context.Context, appID string) ([]siteadmin.Collaborator, error)
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

	collaborators, err := h.Service.ListCollaborators(r.Context(), params.AppID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	response := siteadmin.CollaboratorsListResponse{
		Collaborators: collaborators,
	}
	SiteAdminAPISuccessResponse{Body: response}.WriteTo(w)
}
