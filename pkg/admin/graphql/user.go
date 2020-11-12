package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeUser = "User"

var nodeUser = entity(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeUser,
		Description: "Authgear user",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeUser, nil),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"lastLoginAt": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The last login time of user",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*user.User).LastLoginAt, nil
				},
			},
			"identities": &graphql.Field{
				Type: connIdentity.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*user.User)
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
					source := p.Source.(*user.User)
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
					source := p.Source.(*user.User)
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
					source := p.Source.(*user.User)
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
		},
	}),
	&user.User{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Users.Load(id).Value, nil
	},
)

var connUser = graphqlutil.NewConnectionDef(nodeUser)
