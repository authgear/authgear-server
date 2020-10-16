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
		},
	}),
	&model.User{},
	func(ctx context.Context, id string) (interface{}, error) {
		gqlCtx := GQLContext(ctx)
		lazy := gqlCtx.Viewer.Get()
		return lazy.Map(func(value interface{}) (interface{}, error) {
			user := value.(*model.User)
			if user.ID != id {
				return nil, nil
			}
			return user, nil
		}), nil
	},
)
