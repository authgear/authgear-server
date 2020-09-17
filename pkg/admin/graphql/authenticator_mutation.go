package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
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
		"success": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Boolean),
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
			authenticatorRef, err := decodeAuthenticatorID(resolvedNodeID.ID)
			if err != nil {
				return nil, apierrors.NewInvalid("invalid authenticator ID")
			}

			gqlCtx := GQLContext(p.Context)
			lazy := gqlCtx.Authenticators.Get(authenticatorRef)
			return lazy.
				Map(func(value interface{}) (interface{}, error) {
					i := value.(*authenticator.Info)
					if i == nil {
						return nil, apierrors.NewNotFound("authenticator not found")
					}
					return gqlCtx.Authenticators.Remove(i), nil
				}).
				Map(func(value interface{}) (interface{}, error) {
					return map[string]bool{"success": true}, nil
				}).Value, nil
		},
	},
)
