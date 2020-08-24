package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var query = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node": nodeDefs.NodeField,
		"viewer": &graphql.Field{
			Description: "The current viewer",
			Type:        nodeUser,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				lazy := ctx.Viewer.Get()
				return lazy.Value, nil
			},
		},
		"apps": &graphql.Field{
			Description: "All apps accessible by the viewer",
			Type:        connApp.ConnectionType,
			Args:        relay.ConnectionArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				args := relay.NewConnectionArguments(p.Args)
				gqlCtx := GQLContext(p.Context)
				result, err := gqlCtx.Apps.QueryPage(graphqlutil.NewPageArgs(args))
				if err != nil {
					return nil, err
				}
				return graphqlutil.NewConnection(result), nil
			},
		},
	},
})
