package graphql

import (
	"context"
	"fmt"
	"reflect"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/session"
)

type NodeResolver func(ctx context.Context, id string) (interface{}, error)

var resolvers = map[string]NodeResolver{}
var nodeTypes = map[reflect.Type]*graphql.Object{}

var nodeDefs = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
	IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
		// Access Control: authenticated user.
		sessionInfo := session.GetValidSessionInfo(ctx)
		if sessionInfo == nil {
			return nil, Unauthenticated.New("only authenticated users can query node")
		}
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
		return resolver(ctx, resolvedID.ID)
	},
	TypeResolve: func(params graphql.ResolveTypeParams) *graphql.Object {
		objType, ok := nodeTypes[reflect.TypeOf(params.Value)]
		if !ok {
			panic(fmt.Sprintf("graphql: unknown value type: %T", params.Value))
		}
		return objType
	},
})

func node(schema *graphql.Object, modelType interface{}, resolver NodeResolver) *graphql.Object {
	resolvers[schema.Name()] = resolver
	nodeTypes[reflect.TypeOf(modelType)] = schema
	return schema
}
