package graphql

import (
	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var deleteIdentityInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteIdentityInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"identityID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target identity ID.",
		},
	},
})

var deleteIdentityPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteIdentityPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"deleteIdentity",
	&graphql.Field{
		Description: "Delete identity of user",
		Type:        graphql.NewNonNull(deleteIdentityPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteIdentityInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			identityNodeID := input["identityID"].(string)

			resolvedNodeID := relay.FromGlobalID(identityNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeIdentity {
				return nil, apierrors.NewInvalid("invalid identity ID")
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			info, err := gqlCtx.IdentityFacade.Get(ctx, resolvedNodeID.ID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.IdentityFacade.Remove(ctx, info)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationDeleteIdentityExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: info.UserID,
					},
				},
				Identity: info.ToModel(),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, info.UserID),
			}).Value, nil
		},
	},
)

var createIdentityInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateIdentityInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"definition": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(identityDef),
			Description: "Definition of the new identity.",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Password for the user if required.",
		},
	},
})

var createIdentityPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateIdentityPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
		"identity": &graphql.Field{
			Type: graphql.NewNonNull(nodeIdentity),
		},
	},
})

var _ = registerMutationField(
	"createIdentity",
	&graphql.Field{
		Description: "Create new identity for user",
		Type:        graphql.NewNonNull(createIdentityPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createIdentityInput),
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

			defData := input["definition"].(map[string]interface{})
			identityDef, err := model.ParseIdentityDef(defData)
			if err != nil {
				return nil, err
			}

			password, _ := input["password"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			ref, err := gqlCtx.IdentityFacade.Create(ctx, userID, identityDef, password)
			if err != nil {
				return nil, err
			}

			info, err := gqlCtx.IdentityFacade.Get(ctx, ref.ID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationCreateIdentityExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
				Identity: info.ToModel(),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user":     gqlCtx.Users.Load(ctx, userID),
				"identity": ref,
			}).Value, nil
		},
	},
)

var updateIdentityInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateIdentityInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"identityID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target identity ID.",
		},
		"definition": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(identityDef),
			Description: "New definition of the identity.",
		},
	},
})

var updateIdentityPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateIdentityPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
		"identity": &graphql.Field{
			Type: graphql.NewNonNull(nodeIdentity),
		},
	},
})

var _ = registerMutationField(
	"updateIdentity",
	&graphql.Field{
		Description: "Update an existing identity of user",
		Type:        graphql.NewNonNull(updateIdentityPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateIdentityInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedUserNodeID := relay.FromGlobalID(userNodeID)
			if resolvedUserNodeID == nil || resolvedUserNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedUserNodeID.ID

			identityNodeID := input["identityID"].(string)
			resolvedIdentityNodeID := relay.FromGlobalID(identityNodeID)
			if resolvedIdentityNodeID == nil || resolvedIdentityNodeID.Type != typeIdentity {
				return nil, apierrors.NewInvalid("invalid identity ID")
			}
			identityID := resolvedIdentityNodeID.ID

			defData := input["definition"].(map[string]interface{})
			identityDef, err := model.ParseIdentityDef(defData)
			if err != nil {
				return nil, err
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			ref, err := gqlCtx.IdentityFacade.Update(ctx, identityID, userID, identityDef)
			if err != nil {
				return nil, err
			}

			info, err := gqlCtx.IdentityFacade.Get(ctx, ref.ID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationUpdateIdentityExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: ref.UserID,
					},
				},
				Identity: info.ToModel(),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user":     gqlCtx.Users.Load(ctx, userID),
				"identity": ref,
			}).Value, nil
		},
	},
)
