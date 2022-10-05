package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeUser = "User"

var nodeUser = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeUser,
		Description: "Authgear user",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeUser),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"lastLoginAt": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The last login time of user",
			},
			"identities": &graphql.Field{
				Type: connIdentity.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)
					refs, err := gqlCtx.IdentityFacade.List(source.ID)
					if err != nil {
						return nil, err
					}

					var identities []interface{}
					for _, i := range refs {
						identities = append(identities, i)
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(identities, args), nil
				},
			},
			"authenticators": &graphql.Field{
				Type: connAuthenticator.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)
					refs, err := gqlCtx.AuthenticatorFacade.List(source.ID)
					if err != nil {
						return nil, err
					}

					var authenticators []interface{}
					for _, i := range refs {
						authenticators = append(authenticators, i)
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(authenticators, args), nil

				},
			},
			"verifiedClaims": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(claim))),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)
					claims, err := gqlCtx.VerificationFacade.Get(source.ID)
					if err != nil {
						return nil, err
					}

					return claims, nil
				},
			},
			"sessions": &graphql.Field{
				Type: connSession.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)
					ss, err := gqlCtx.SessionFacade.List(source.ID)
					if err != nil {
						return nil, err
					}

					var sessions []interface{}
					for _, i := range ss {
						sessions = append(sessions, i.ToAPIModel())
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(sessions, args), nil
				},
			},
			"authorizations": &graphql.Field{
				Type: connAuthorization.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)

					// return third party client authorizations only in admin api
					filter := oauth.NewKeepThirdPartyAuthorizationFilter(gqlCtx.OAuthConfig)
					as, err := gqlCtx.AuthorizationFacade.List(source.ID, filter)
					if err != nil {
						return nil, err
					}

					var authzs []interface{}
					for _, i := range as {
						authzs = append(authzs, i.ToAPIModel())
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(authzs, args), nil
				},
			},
			"isDisabled": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"isAnonymous": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"disableReason": &graphql.Field{
				Type: graphql.String,
			},
			"isDeactivated": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"deleteAt": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The scheduled deletion time of the user",
			},
			"standardAttributes": &graphql.Field{
				Type: graphql.NewNonNull(UserStandardAttributes),
			},
			"customAttributes": &graphql.Field{
				Type: graphql.NewNonNull(UserCustomAttributes),
			},
			"web3": &graphql.Field{
				Type: graphql.NewNonNull(Web3Claims),
			},
			"formattedName": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					formattedName := stdattrs.T(source.StandardAttributes).FormattedName()
					if formattedName == "" {
						return nil, nil
					}
					return formattedName, nil
				},
			},
			"endUserAccountID": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					endUserAccountID := source.EndUserAccountID()
					if endUserAccountID == "" {
						return nil, nil
					}
					return endUserAccountID, nil
				},
			},
		},
	}),
	&model.User{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Users.Load(id).Value, nil
	},
)

var connUser = graphqlutil.NewConnectionDef(nodeUser)
