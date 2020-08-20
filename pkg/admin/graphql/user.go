package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
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
					ref := p.Source.(*user.User)
					identities := GQLContext(p.Context).Identities.List(ref.ID)
					result := identities.Map(func(value interface{}) (interface{}, error) {
						var identities []interface{}
						for _, i := range value.([]*identity.Ref) {
							identities = append(identities, i)
						}
						args := relay.NewConnectionArguments(p.Args)
						return NewConnectionFromArray(identities, args), nil
					})
					return result.Value, nil
				},
			},
			"authenticators": &graphql.Field{
				Type: connAuthenticator.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ref := p.Source.(*user.User)
					authenticators := GQLContext(p.Context).Authenticators.List(ref.ID)
					result := authenticators.Map(func(value interface{}) (interface{}, error) {
						var authenticators []interface{}
						for _, i := range value.([]*authenticator.Ref) {
							authenticators = append(authenticators, i)
						}
						args := relay.NewConnectionArguments(p.Args)
						return NewConnectionFromArray(authenticators, args), nil
					})
					return result.Value, nil
				},
			},
		},
	}),
	&user.User{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Users.Get(id).Value, nil
	},
)

var connUser = connection(nodeUser)
