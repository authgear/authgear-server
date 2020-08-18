package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/admin/loader"
)

var query = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node": nodeDefs.NodeField,
		"users": &graphql.Field{
			Description: "All users",
			Type:        connUser.ConnectionType,
			Args:        relay.ConnectionArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				args := relay.NewConnectionArguments(p.Args)
				result, err := GQLContext(p.Context).Users.QueryPage(loader.NewPageArgs(args))
				if err != nil {
					return nil, err
				}
				return NewConnection(result), nil
			},
		},
	},
})
