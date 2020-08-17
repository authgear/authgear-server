package graphql

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

var nodeDefs = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
	IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
		panic("not implemented")
	},
	TypeResolve: func(p graphql.ResolveTypeParams) *graphql.Object {
		panic("not implemented")
	},
})
