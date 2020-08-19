package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/lib/api/model"
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
			"id":        relay.GlobalIDField(typeUser, nil),
			"createdAt": entityCreatedAtField,
			"updatedAt": entityUpdatedAtField,
		},
	}),
	&model.User{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Users.Get(id)
	},
)

var connUser = connection(nodeUser)
