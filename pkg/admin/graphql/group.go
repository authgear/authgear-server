package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeGroup = "Group"

var nodeGroup = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeGroup,
		Description: "Authgear group",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeGroup),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"key": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The key of the group.",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional name of the group.",
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional description of the group.",
			},
		},
	}),
	&model.Group{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Groups.Load(id).Value, nil
	},
)

var connGroup = graphqlutil.NewConnectionDef(nodeGroup)
