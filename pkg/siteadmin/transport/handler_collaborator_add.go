package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureCollaboratorAddRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/api/v1/projects/:projectID/collaborators")
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

type CollaboratorAddHandler struct {
	// Add service dependencies here as needed
}

type CollaboratorAddParams struct {
	ProjectID string
	UserEmail string
}

func parseCollaboratorAddParams(r *http.Request) (CollaboratorAddParams, error) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return CollaboratorAddParams{}, err
	}

	if err := CollaboratorAddRequestSchema.Validator().ValidateValue(r.Context(), body); err != nil {
		return CollaboratorAddParams{}, err
	}

	return CollaboratorAddParams{
		ProjectID: httproute.GetParam(r, "projectID"),
		UserEmail: body["user_email"].(string),
	}, nil
}

func (h *CollaboratorAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseCollaboratorAddParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	_ = params

	// TODO: implement
	http.NotFound(w, r)
}
