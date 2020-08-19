package graphql

import (
	"github.com/graphql-go/graphql"
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
				return ctx.Viewer.Get()
			},
		},
	},
})
