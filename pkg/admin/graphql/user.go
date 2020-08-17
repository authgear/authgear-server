package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

var userNode = graphql.NewObject(graphql.ObjectConfig{
	Name:        "User",
	Description: "Authgear user",
	Interfaces: []*graphql.Interface{
		nodeDefs.NodeInterface,
	},
	Fields: graphql.Fields{
		"id": relay.GlobalIDField("User", nil),
	},
})

var userConn = relay.ConnectionDefinitions(relay.ConnectionConfig{
	Name:     "User",
	NodeType: userNode,
})
