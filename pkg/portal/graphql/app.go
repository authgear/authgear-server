package graphql

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeApp = "App"

var nodeApp = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeApp,
		Description: "Authgear app",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": relay.GlobalIDField(typeApp, nil),
			"appConfig": &graphql.Field{
				Type: graphql.NewNonNull(graphqlutil.JSONObject),
			},
			"secretConfig": &graphql.Field{
				Type: graphql.NewNonNull(graphqlutil.JSONObject),
			},
		},
	}),
	&model.App{},
	func(ctx context.Context, id string) (interface{}, error) {
		gqlCtx := GQLContext(ctx)
		lazy := gqlCtx.Apps.Get(id)
		return lazy.Value, nil
	},
)

var connApp = connection(nodeApp)
