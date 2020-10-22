package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var deleteAuthenticatorInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteAuthenticatorInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"authenticatorID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target authenticator ID.",
		},
	},
})

var deleteAuthenticatorPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteAuthenticatorPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"deleteAuthenticator",
	&graphql.Field{
		Description: "Delete authenticator of user",
		Type:        graphql.NewNonNull(deleteAuthenticatorPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteAuthenticatorInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			authenticatorNodeID := input["authenticatorID"].(string)

			resolvedNodeID := relay.FromGlobalID(authenticatorNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeAuthenticator {
				return nil, apierrors.NewInvalid("invalid authenticator ID")
			}
			authenticatorRef, err := loader.DecodeAuthenticatorID(resolvedNodeID.ID)
			if err != nil {
				return nil, apierrors.NewInvalid("invalid authenticator ID")
			}

			gqlCtx := GQLContext(p.Context)

			info, err := gqlCtx.AuthenticatorFacade.Get(authenticatorRef)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuthenticatorFacade.Remove(info)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(info.UserID),
			}).Value, nil
		},
	},
)
