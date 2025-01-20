package graphql

import (
	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
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

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			s, err := gqlCtx.SessionFacade.Get(ctx, resolvedNodeID.ID)
			if err != nil {
				return nil, err
			}
			userID := s.GetAuthenticationInfo().UserID

			err = gqlCtx.SessionFacade.Revoke(ctx, s.SessionID())
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationRevokeSessionExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
				Session: *s.ToAPIModel(),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
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

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.SessionFacade.RevokeAll(ctx, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationRevokeAllSessionsExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var createSessionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateSessionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"clientID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Target client ID.",
		},
	},
})

var createSessionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateSessionPayload",
	Fields: graphql.Fields{
		"refreshToken": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"accessToken": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var _ = registerMutationField(
	"createSession",
	&graphql.Field{
		Description: "Create a session of a given user",
		Type:        graphql.NewNonNull(createSessionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createSessionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			if !(*gqlCtx.AdminAPIFeatureConfig.CreateSessionEnabled) {
				return nil, apierrors.NewForbidden("CreateSession is disabled")
			}

			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			clientID := input["clientID"].(string)

			s, resp, err := gqlCtx.OAuthFacade.CreateSession(ctx, clientID, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationCreateSessionExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
				Session: *s.ToAPIModel(),
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"refreshToken": resp["refresh_token"],
				"accessToken":  resp["access_token"],
			}, nil
		},
	},
)
