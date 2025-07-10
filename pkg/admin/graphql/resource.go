package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/graphql-go/graphql"
)

const typeResource = "Resource"

var ErrInvalidResourceID = apierrors.NewInvalid("invalid resource ID")

var nodeResource = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeResource,
		Description: "Authgear resource",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeResource),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"uri": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The URI of the resource.",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional name of the resource.",
			},
		},
	}),
	&model.Resource{},
	func(ctx context.Context, gqlCtx *Context, id string) (interface{}, error) {
		return gqlCtx.Resources.Load(ctx, id).Value, nil
	},
)

var connResource = graphqlutil.NewConnectionDef(nodeResource)
