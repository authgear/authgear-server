package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var revokeSessionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RevokeSessionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"sessionID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target session ID.",
		},
	},
})

var revokeSessionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RevokeSessionPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"revokeSession",
	&graphql.Field{
		Description: "Revoke session of user",
		Type:        graphql.NewNonNull(revokeSessionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(revokeSessionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			sessionID := input["sessionID"].(string)

			resolvedNodeID := relay.FromGlobalID(sessionID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeSession {
				return nil, apierrors.NewInvalid("invalid session ID")
			}

			gqlCtx := GQLContext(p.Context)

			s, err := gqlCtx.SessionFacade.Get(resolvedNodeID.ID)
			if err != nil {
				return nil, err
			}
			userID := s.GetAuthenticationInfo().UserID

			err = gqlCtx.SessionFacade.Revoke(s.SessionID())
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil
		},
	},
)

var revokeAllSessionsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RevokeAllSessionsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var revokeAllSessionsPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RevokeAllSessionsPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"revokeAllSessions",
	&graphql.Field{
		Description: "Revoke all sessions of user",
		Type:        graphql.NewNonNull(revokeAllSessionsPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(revokeAllSessionsInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			gqlCtx := GQLContext(p.Context)

			err := gqlCtx.SessionFacade.RevokeAll(userID)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil
		},
	},
)
