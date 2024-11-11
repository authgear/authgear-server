package graphql

import (
	"context"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeAuthorization = "Authorization"

var nodeAuthorization = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name: typeAuthorization,
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeAuthorization),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"clientID": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"scopes": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			},
		},
	}),
	&model.Authorization{},
	func(ctx context.Context, gqlCtx *Context, id string) (interface{}, error) {
		authz, err := gqlCtx.AuthorizationFacade.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		return authz.ToAPIModel(), nil
	},
)

var connAuthorization = graphqlutil.NewConnectionDef(nodeAuthorization)
