package graphql

import (
	"context"
	"fmt"
	"reflect"

	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"
)

type Resolver func(ctx *Context, id string) (interface{}, error)

var resolvers = map[string]Resolver{}
var typeMapping = map[reflect.Type]*graphql.Object{}

var nodeDefs = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
	IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
		// If the ID is invalid, we should return null instead of returning an error.
		// This behavior conforms the schema.
		resolvedID := relay.FromGlobalID(id)
		if resolvedID == nil {
			return nil, nil
		}
		resolver, ok := resolvers[resolvedID.Type]
		if !ok {
			return nil, nil
		}
		return resolver(GQLContext(ctx), resolvedID.ID)
	},
	TypeResolve: func(params graphql.ResolveTypeParams) *graphql.Object {
		objType, ok := typeMapping[reflect.TypeOf(params.Value)]
		if !ok {
			panic(fmt.Sprintf("graphql: unknown value type: %T", params.Value))
		}
		return objType
	},
})

func node(schema *graphql.Object, modelType interface{}, resolver Resolver) *graphql.Object {
	resolvers[schema.Name()] = resolver
	typeMapping[reflect.TypeOf(modelType)] = schema
	return schema
}
