package graphql

import (
	"github.com/graphql-go/graphql"

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
		},
	}),
	&user.User{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Users.Get(id).Value, nil
	},
)

var connUser = connection(nodeUser)
