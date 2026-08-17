package graphql

import (
	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/dcr"
)

var createInitialAccessTokenInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateInitialAccessTokenInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"expiresIn": &graphql.InputObjectFieldConfig{
			Type:        graphql.Int,
			Description: "Token lifetime in seconds. If omitted, a server default is used.",
		},
		"type": &graphql.InputObjectFieldConfig{
			Type:        initialAccessTokenType,
			Description: "Defaults to THIRD_PARTY.",
		},
	},
})

var createInitialAccessTokenPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateInitialAccessTokenPayload",
	Fields: graphql.Fields{
		"token": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The opaque IAT value. Returned ONCE only.",
		},
		"initialAccessToken": &graphql.Field{
			Type: graphql.NewNonNull(nodeInitialAccessToken),
		},
	},
})

var _ = registerMutationField(
	"createInitialAccessToken",
	&graphql.Field{
		Description: "Creates an opaque Initial Access Token for use with POST /oauth2/register.",
		Type:        graphql.NewNonNull(createInitialAccessTokenPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createInitialAccessTokenInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			input := p.Args["input"].(map[string]any)

			var expiresIn *int
			if v, ok := input["expiresIn"].(int); ok {
				expiresIn = &v
			}
			iatType := dcr.InitialAccessTokenTypeThirdParty
			if v, ok := input["type"].(string); ok {
				iatType = dcr.InitialAccessTokenType(v)
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			token, iat, err := gqlCtx.DCRFacade.CreateInitialAccessToken(ctx, &dcr.NewInitialAccessTokenOptions{
				ExpiresIn: expiresIn,
				Type:      iatType,
			})
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"token":              token,
				"initialAccessToken": iat,
			}, nil
		},
	},
)

var revokeInitialAccessTokenInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RevokeInitialAccessTokenInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target initial access token ID.",
		},
	},
})

var revokeInitialAccessTokenPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RevokeInitialAccessTokenPayload",
	Fields: graphql.Fields{
		"ok": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var _ = registerMutationField(
	"revokeInitialAccessToken",
	&graphql.Field{
		Description: "Revokes an Initial Access Token so it can no longer be used for registration.",
		Type:        graphql.NewNonNull(revokeInitialAccessTokenPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(revokeInitialAccessTokenInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			input := p.Args["input"].(map[string]any)

			resolvedNodeID := relay.FromGlobalID(input["id"].(string))
			if resolvedNodeID == nil || resolvedNodeID.Type != typeInitialAccessToken {
				return nil, apierrors.NewInvalid("invalid initial access token ID")
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			if err := gqlCtx.DCRFacade.RevokeInitialAccessToken(ctx, resolvedNodeID.ID); err != nil {
				return nil, err
			}

			return map[string]any{"ok": true}, nil
		},
	},
)
