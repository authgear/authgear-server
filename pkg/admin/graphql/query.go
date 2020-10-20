package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var query = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node":  nodeDefs.NodeField,
		"nodes": nodeDefs.NodesField,
		"users": &graphql.Field{
			Description: "All users",
			Type:        connUser.ConnectionType,
			Args:        relay.ConnectionArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				args := relay.NewConnectionArguments(p.Args)
				gqlCtx := GQLContext(p.Context)
				result, err := gqlCtx.UserFacade.QueryPage(graphqlutil.NewPageArgs(args))
				if err != nil {
					return nil, err
				}
				return graphqlutil.NewConnection(result), nil
			},
		},
	},
})
