package graphql

import (
	"context"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeScope = "Scope"

var ErrInvalidScopeID = apierrors.NewInvalid("invalid scope ID")

var nodeScope = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeScope,
		Description: "Authgear scope",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeScope),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"resourceID": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.Scope)
					return relay.ToGlobalID(typeResource, source.ResourceID), nil
				},
				Description: "The resource ID.",
			},
			"scope": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The scope string.",
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional description of the scope.",
			},
		},
	}),
	&model.Scope{},
	func(ctx context.Context, gqlCtx *Context, id string) (interface{}, error) {
		return gqlCtx.Scopes.Load(ctx, id).Value, nil
	},
)

var connScope = graphqlutil.NewConnectionDef(nodeScope)
