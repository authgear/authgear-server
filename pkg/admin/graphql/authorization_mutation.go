package graphql

import (
	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var deleteAuthorizationInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteAuthorizationInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"authorizationID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target authorization ID.",
		},
	},
})

var deleteAuthorizationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteAuthorizationPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"deleteAuthorization",
	&graphql.Field{
		Description: "Delete authorization",
		Type:        graphql.NewNonNull(deleteAuthorizationPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteAuthorizationInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			authorizationID := input["authorizationID"].(string)

			resolvedNodeID := relay.FromGlobalID(authorizationID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeAuthorization {
				return nil, apierrors.NewInvalid("invalid authorization ID")
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			authz, err := gqlCtx.AuthorizationFacade.Get(ctx, resolvedNodeID.ID)
			if err != nil {
				return nil, err
			}
			userID := authz.UserID

			err = gqlCtx.AuthorizationFacade.Delete(ctx, authz)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationDeleteAuthorizationExecutedEventPayload{
				Authorization: *authz.ToAPIModel(),
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
