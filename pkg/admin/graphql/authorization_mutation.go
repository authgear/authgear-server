package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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

			gqlCtx := GQLContext(p.Context)

			authz, err := gqlCtx.AuthorizationFacade.Get(resolvedNodeID.ID)
			if err != nil {
				return nil, err
			}
			userID := authz.UserID

			err = gqlCtx.AuthorizationFacade.Delete(authz)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil
		},
	},
)
