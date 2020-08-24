package graphqlutil

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

type Connection struct {
	Edges      []*relay.Edge  `json:"edges"`
	PageInfo   relay.PageInfo `json:"pageInfo"`
	TotalCount interface{}    `json:"totalCount"`
}

func NewConnection(result *PageResult) *Connection {
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
		TotalCount: result.TotalCount.Value,
	}
}

func NewConnectionFromArray(data []interface{}, args relay.ConnectionArguments) *Connection {
	conn := relay.ConnectionFromArray(data, args)
	return &Connection{
		Edges:      conn.Edges,
		PageInfo:   conn.PageInfo,
		TotalCount: len(data),
	}
}

func NewConnectionDef(schema *graphql.Object) *relay.GraphQLConnectionDefinitions {
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
