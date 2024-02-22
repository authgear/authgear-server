package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
)

const typeRole = "Role"

var nodeRole = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeRole,
		Description: "Authgear role",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeRole),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"key": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The key of the role.",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional name of the role.",
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional description of the role.",
			},
		},
	}),
	&model.Role{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Roles.Load(id).Value, nil
	},
)
