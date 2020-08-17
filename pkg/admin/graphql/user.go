package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/lib/api/model"
)

const typeUser = "User"

var nodeUser = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeUser,
		Description: "Authgear user",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": relay.GlobalIDField(typeUser, nil),
		},
	}),
	&model.User{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Users.Get(id)
	},
)

var connUser = connection(nodeUser)
