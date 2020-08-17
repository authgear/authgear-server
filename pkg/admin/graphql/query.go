package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

var query = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node": nodeDefs.NodeField,
		"users": &graphql.Field{
			Description: "All users",
			Type:        userConn.ConnectionType,
			Args:        relay.ConnectionArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// TODO: implement this
				args := relay.NewConnectionArguments(p.Args)
				var users []interface{}
				return relay.ConnectionFromArray(users, args), nil
			},
		},
	},
})
