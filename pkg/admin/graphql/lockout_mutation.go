package graphql

import (
	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
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
		Resolve: func(p graphql.ResolveParams) (any, error) {
			input := p.Args["input"].(map[string]any)
			userNodeID := input["userID"].(string)

			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			previousStatus, err := gqlCtx.AccountLockoutFacade.GetAccountLockoutStatus(ctx, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AccountLockoutFacade.ResetAccountLockout(ctx, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationResetAccountLockoutExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
				PreviousLockoutStatus: previousStatus,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]any{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)
