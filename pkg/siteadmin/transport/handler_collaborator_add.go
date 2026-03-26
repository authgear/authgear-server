package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureCollaboratorAddRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").
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

// TODO: Replace dummy data with real implementation.
var dummyCollaborators = map[string][]siteadmin.Collaborator{
	"project-alpha": {
		{
			Id:        "collab-1",
			ProjectId: "project-alpha",
			UserId:    "user-001",
			UserEmail: "alice@example.com",
			Role:      siteadmin.Owner,
			CreatedAt: time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
		},
		{
			Id:        "collab-2",
			ProjectId: "project-alpha",
			UserId:    "user-002",
			UserEmail: "bob@example.com",
			Role:      siteadmin.Editor,
			CreatedAt: time.Date(2024, 2, 10, 9, 0, 0, 0, time.UTC),
		},
	},
	"project-beta": {
		{
			Id:        "collab-3",
			ProjectId: "project-beta",
			UserId:    "user-003",
			UserEmail: "carol@example.com",
			Role:      siteadmin.Owner,
			CreatedAt: time.Date(2024, 3, 22, 10, 0, 0, 0, time.UTC),
		},
	},
}

func dummyCollaboratorsForProject(projectID string) []siteadmin.Collaborator {
	if collaborators, ok := dummyCollaborators[projectID]; ok {
		return collaborators
	}
	return []siteadmin.Collaborator{}
}

type CollaboratorAddHandler struct {
	// Add service dependencies here as needed
}

type CollaboratorAddParams struct {
	ProjectID string
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
		ProjectID:              httproute.GetParam(r, "projectID"),
		AddCollaboratorRequest: body,
	}, nil
}

func (h *CollaboratorAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseCollaboratorAddParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	// TODO: Replace with real data source. Return a dummy collaborator now.
	collaborator := siteadmin.Collaborator{
		Id:        "collab-new",
		ProjectId: params.ProjectID,
		UserId:    "user-new",
		UserEmail: params.UserEmail,
		Role:      siteadmin.Editor,
		CreatedAt: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(collaborator)
}
