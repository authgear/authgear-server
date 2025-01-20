package graphqlutil

import (
	"reflect"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/graphqlgo/relay"
)

type LazyItem struct {
	Lazy   *Lazy
	Cursor Cursor
}

type Connection struct {
	Edges      []*relay.Edge  `json:"edges"`
	PageInfo   relay.PageInfo `json:"pageInfo"`
	TotalCount interface{}    `json:"totalCount"`
}

func NewConnectionFromResult(lazyItems []LazyItem, result *PageResult) (*Connection, error) {
	var edges = make([]*relay.Edge, 0, len(lazyItems))
	for _, item := range lazyItems {
		node, err := item.Lazy.Value()
		if err != nil {
			return nil, err
		}
		// I have no idea why node != nil is false even node is really nil.
		// We have to use reflect and use IsNil to check nil.
		nodeValue := reflect.ValueOf(node)
		if !nodeValue.IsNil() {
			edges = append(edges, &relay.Edge{
				Node:   node,
				Cursor: relay.ConnectionCursor(item.Cursor),
			})
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
	}, nil
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
