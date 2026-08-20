package graphql

import (
	"context"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
)

const typeInitialAccessToken = "InitialAccessToken"

var initialAccessTokenType = graphql.NewEnum(graphql.EnumConfig{
	Name: "InitialAccessTokenType",
	Values: graphql.EnumValueConfigMap{
		"THIRD_PARTY": &graphql.EnumValueConfig{
			Value: string(model.OAuthInitialAccessTokenTypeThirdParty),
		},
		"FIRST_PARTY": &graphql.EnumValueConfig{
			Value: string(model.OAuthInitialAccessTokenTypeFirstParty),
		},
	},
})

var nodeInitialAccessToken = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeInitialAccessToken,
		Description: "Initial Access Token for Dynamic Client Registration",
		Interfaces: []*graphql.Interface{
			// NOT entityInterface: the spec's InitialAccessToken type has no
			// updatedAt field. This is independent of model.Meta being
			// embedded on model.OAuthInitialAccessToken: entityIDField and
			// entityCreatedAtField below require Meta at runtime regardless
			// of which interfaces the object declares.
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeInitialAccessToken),
			"createdAt": entityCreatedAtField(nil),
			"expiresAt": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.DateTime),
				Description: "The expiry time of the initial access token.",
			},
			"type": &graphql.Field{
				Type:        graphql.NewNonNull(initialAccessTokenType),
				Description: "The type of the initial access token.",
				Resolve: func(p graphql.ResolveParams) (any, error) {
					iat := p.Source.(*model.OAuthInitialAccessToken)
					return string(iat.Type), nil
				},
			},
		},
	}),
	&model.OAuthInitialAccessToken{},
	func(ctx context.Context, gqlCtx *Context, id string) (any, error) {
		return gqlCtx.InitialAccessTokens.Load(ctx, id).Value, nil
	},
)
