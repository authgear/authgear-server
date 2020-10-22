package graphql

import (
	"context"

	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

const typeUser = "User"

var nodeUser = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeUser,
		Description: "Portal User",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": relay.GlobalIDField(typeUser, nil),
			"email": &graphql.Field{
				Type: graphql.String,
			},
		},
	}),
	&model.User{},
	func(ctx context.Context, id string) (interface{}, error) {
		// FIXME(portal): How can we determine if the viewer can access this user?
		gqlCtx := GQLContext(ctx)
		return gqlCtx.Users.Load(id).Value, nil
	},
)
