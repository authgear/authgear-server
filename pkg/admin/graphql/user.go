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
			"id": globalIDField(typeUser, func(obj interface{}) (string, error) {
				return obj.(*user.User).ID, nil
			}),
			"createdAt": entityCreatedAtField,
			"updatedAt": entityUpdatedAtField,
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
		thunk := ctx.Users.Get(id)
		return func() (interface{}, error) {
			return thunk()
		}, nil
	},
)

var connUser = connection(nodeUser)
