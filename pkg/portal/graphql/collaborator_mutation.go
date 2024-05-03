package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var deleteCollaboratorInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteCollaboratorInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"collaboratorID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Collaborator ID.",
		},
	},
})

var deleteCollaboratorPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteCollaboratorPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"deleteCollaborator",
	&graphql.Field{
		Description: "Delete collaborator of target app.",
		Type:        graphql.NewNonNull(deleteCollaboratorPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteCollaboratorInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can delete collaborator")
			}

			input := p.Args["input"].(map[string]interface{})
			collaboratorID := input["collaboratorID"].(string)

			gqlCtx := GQLContext(p.Context)

			targetCollab, err := gqlCtx.CollaboratorService.GetCollaborator(collaboratorID)
			if err != nil {
				return nil, err
			}

			appID := targetCollab.AppID

			// Access Control: collaborator.
			userID, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			selfCollab, err := gqlCtx.CollaboratorService.GetCollaboratorByAppAndUser(appID, userID)
			if err != nil {
				return nil, err
			}
			if selfCollab.Role.Level() > targetCollab.Role.Level() {
				return nil, AccessDenied.Errorf("insufficient permission to delete %s collaborators", targetCollab.Role)
			}

			err = gqlCtx.CollaboratorService.DeleteCollaborator(targetCollab)
			if err != nil {
				return nil, err
			}

			app, err := gqlCtx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectCollaboratorDeletedEventPayload{
				CollaboratorID:     targetCollab.ID,
				CollaboratorUserID: targetCollab.UserID,
				CollaboratorRole:   string(targetCollab.Role),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": gqlCtx.Apps.Load(appID),
			}).Value, nil
		},
	},
)

var deleteCollaboratorInvitationInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteCollaboratorInvitationInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"collaboratorInvitationID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Collaborator invitation ID.",
		},
	},
})

var deleteCollaboratorInvitationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteCollaboratorInvitationPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"deleteCollaboratorInvitation",
	&graphql.Field{
		Description: "Delete collaborator invitation of target app.",
		Type:        graphql.NewNonNull(deleteCollaboratorInvitationPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteCollaboratorInvitationInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can delete collaborator invitation")
			}

			input := p.Args["input"].(map[string]interface{})
			collaboratorInvitationID := input["collaboratorInvitationID"].(string)

			gqlCtx := GQLContext(p.Context)

			invitation, err := gqlCtx.CollaboratorService.GetInvitation(collaboratorInvitationID)
			if err != nil {
				return nil, err
			}

			// Access Control: collaborator.
			_, err = gqlCtx.AuthzService.CheckAccessOfViewer(invitation.AppID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.CollaboratorService.DeleteInvitation(invitation)
			if err != nil {
				return nil, err
			}

			app, err := gqlCtx.AppService.Get(invitation.AppID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectCollaboratorInvitationDeletedEventPayload{
				InviteeEmail: invitation.InviteeEmail,
				InvitedBy:    invitation.InvitedBy,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": gqlCtx.Apps.Load(invitation.AppID),
			}).Value, nil
		},
	},
)

var createCollaboratorInvitationInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateCollaboratorInvitationInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target app ID.",
		},
		"inviteeEmail": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Invitee email address.",
		},
	},
})

var createCollaboratorInvitationInputSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"appID": { "type": "string" },
			"inviteeEmail": {
				"type": "string",
				"format": "email"
			}
		},
		"required": ["appID", "inviteeEmail"]
	}
`)

var createCollaboratorInvitationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateCollaboratorInvitationPayload",
	Fields: graphql.Fields{
		"app":                    &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
		"collaboratorInvitation": &graphql.Field{Type: graphql.NewNonNull(collaboratorInvitation)},
	},
})

var _ = registerMutationField(
	"createCollaboratorInvitation",
	&graphql.Field{
		Description: "Invite a collaborator to the target app.",
		Type:        graphql.NewNonNull(createCollaboratorInvitationPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createCollaboratorInvitationInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can create collaborator invitation")
			}

			err := createCollaboratorInvitationInputSchema.Validator().ValidateValue(input)
			if err != nil {
				return nil, err
			}

			appNodeID := input["appID"].(string)
			inviteeEmail := input["inviteeEmail"].(string)

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			gqlCtx := GQLContext(p.Context)

			// Access Control: collaborator.
			_, err = gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			invitation, err := gqlCtx.CollaboratorService.SendInvitation(appID, inviteeEmail)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.TutorialService.RecordProgresses(appID, []tutorial.Progress{tutorial.ProgressInvite})
			if err != nil {
				return nil, err
			}

			gqlCtx.CollaboratorInvitations.Prime(invitation.ID, invitation)

			app, err := gqlCtx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectCollaboratorInvitationCreatedEventPayload{
				InviteeEmail: invitation.InviteeEmail,
				InvitedBy:    invitation.InvitedBy,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app":                    gqlCtx.Apps.Load(appID),
				"collaboratorInvitation": gqlCtx.CollaboratorInvitations.Load(invitation.ID),
			}).Value, nil
		},
	},
)

var acceptCollaboratorInvitationInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AcceptCollaboratorInvitationInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"code": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Invitation code.",
		},
	},
})

var acceptCollaboratorInvitationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AcceptCollaboratorInvitationPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"acceptCollaboratorInvitation",
	&graphql.Field{
		Description: "Accept collaborator invitation to the target app.",
		Type:        graphql.NewNonNull(acceptCollaboratorInvitationPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(acceptCollaboratorInvitationInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			code := input["code"].(string)

			gqlCtx := GQLContext(p.Context)

			// Access Control: authenicated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can accept invitations")
			}

			collaborator, err := gqlCtx.CollaboratorService.AcceptInvitation(code)
			if err != nil {
				return nil, err
			}

			appID := collaborator.AppID

			app, err := gqlCtx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectCollaboratorInvitationAcceptedEventPayload{
				CollaboratorUserID: collaborator.UserID,
				CollaboratorRole:   string(collaborator.Role),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": gqlCtx.Apps.Load(collaborator.AppID),
			}).Value, nil
		},
	},
)
