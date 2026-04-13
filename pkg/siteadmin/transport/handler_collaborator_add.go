package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureCollaboratorAddRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/v1/apps/:appID/collaborators")
}

var CollaboratorAddRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"user_email": { "type": "string", "format": "email" }
		},
		"required": ["user_email"]
	}
`)

type CollaboratorAddService interface {
	AddCollaborator(ctx context.Context, appID string, userEmail string) (*siteadmin.Collaborator, error)
}

type CollaboratorAddHandler struct {
	Service CollaboratorAddService
}

type CollaboratorAddParams struct {
	AppID string
	siteadmin.AddCollaboratorRequest
}

func parseCollaboratorAddParams(r *http.Request) (CollaboratorAddParams, error) {
	var body siteadmin.AddCollaboratorRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return CollaboratorAddParams{}, err
	}

	if err := CollaboratorAddRequestSchema.Validator().ValidateValue(r.Context(), body); err != nil {
		return CollaboratorAddParams{}, err
	}

	return CollaboratorAddParams{
		AppID:                  httproute.GetParam(r, "appID"),
		AddCollaboratorRequest: body,
	}, nil
}

func (h *CollaboratorAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseCollaboratorAddParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	collaborator, err := h.Service.AddCollaborator(r.Context(), params.AppID, params.UserEmail)
	if err != nil {
		writeError(w, r, err)
		return
	}

	SiteAdminAPISuccessResponse{Body: collaborator}.WriteTo(w)
}
