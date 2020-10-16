package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
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
				ctx := GQLContext(p.Context)
				viewer := ctx.Viewer.Get()
				apps := viewer.Map(func(u interface{}) (interface{}, error) {
					userID := u.(*model.User).ID
					return ctx.Apps.List(userID), nil
				})
				result := apps.Map(func(value interface{}) (interface{}, error) {
					var apps []interface{}
					for _, i := range value.([]*model.App) {
						apps = append(apps, i)
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(apps, args), nil
				})
				return result.Value, nil
			},
		},
	},
})
