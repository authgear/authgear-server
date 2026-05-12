package graphql

import (
	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var resetAccountLockoutInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ResetAccountLockoutInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var resetAccountLockoutPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "ResetAccountLockoutPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"resetAccountLockout",
	&graphql.Field{
		Description: "Reset the account lockout state of a user",
		Type:        graphql.NewNonNull(resetAccountLockoutPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(resetAccountLockoutInput),
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

			err := gqlCtx.AccountLockoutFacade.ResetAccountLockout(ctx, userID)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)
