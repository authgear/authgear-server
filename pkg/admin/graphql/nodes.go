package graphql

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/admin/loader"
)

var nodeDefs = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
	IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
		resolvedID := relay.FromGlobalID(id)
		if resolvedID == nil {
			return nil, errors.New("invalid ID")
		}
		resolver, ok := resolvers[resolvedID.Type]
		if !ok {
			return nil, fmt.Errorf("unknown node type: %s", resolvedID.Type)
		}
		return resolver(GQLContext(ctx), resolvedID.ID)
	},
	TypeResolve: func(params graphql.ResolveTypeParams) *graphql.Object {
		objType, ok := nodeTypes[reflect.TypeOf(params.Value)]
		if !ok {
			panic(fmt.Sprintf("graphql: unknown value type: %T", params.Value))
		}
		return objType
	},
})

type NodeResolver func(ctx *Context, id string) (interface{}, error)

var resolvers = map[string]NodeResolver{}
var nodeTypes = map[reflect.Type]*graphql.Object{}

func node(schema *graphql.Object, modelType interface{}, resolver NodeResolver) *graphql.Object {
	resolvers[schema.Name()] = resolver
	nodeTypes[reflect.TypeOf(modelType)] = schema
	return schema
}

type Connection struct {
	Edges      []*relay.Edge  `json:"edges"`
	PageInfo   relay.PageInfo `json:"pageInfo"`
	TotalCount uint64         `json:"totalCount"`
}

func NewConnection(result *loader.PageResult) *Connection {
	var edges = make([]*relay.Edge, len(result.Values))
	for i, item := range result.Values {
		edges[i] = &relay.Edge{
			Node:   item.Value,
			Cursor: relay.ConnectionCursor(item.Cursor),
		}
	}
	pageInfo := relay.PageInfo{
		StartCursor:     "",
		EndCursor:       "",
		HasPreviousPage: result.HasPreviousPage,
		HasNextPage:     result.HasNextPage,
	}
	if len(edges) > 0 {
		pageInfo.StartCursor = edges[0].Cursor
		pageInfo.EndCursor = edges[len(edges)-1].Cursor
	}
	return &Connection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: result.TotalCount,
	}
}

func connection(schema *graphql.Object) *relay.GraphQLConnectionDefinitions {
	return relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     schema.Name(),
		NodeType: schema,
		ConnectionFields: graphql.Fields{
			"totalCount": &graphql.Field{
				Type:        graphql.Int,
				Description: "Total number of nodes in the connection.",
			},
		},
	})
}
