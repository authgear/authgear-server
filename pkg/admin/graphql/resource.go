package graphql

import (
	"context"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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

func init() {
	nodeResource.AddFieldConfig("scopes", &graphql.Field{
		Type:        connScope.ConnectionType,
		Description: "The list of scopes for this resource.",
		Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
			"clientID":      &graphql.ArgumentConfig{Type: graphql.String},
			"searchKeyword": &graphql.ArgumentConfig{Type: graphql.String},
		}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.Resource)
			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			// TODO(tung): Support client ID & searchKeyword filter
			scopes, err := gqlCtx.ResourceScopeFacade.ListScopes(ctx, source.ID)
			if err != nil {
				return nil, err
			}

			scopeIfaces := make([]interface{}, len(scopes))
			for i, s := range scopes {
				scopeIfaces[i] = s
			}

			args := relay.NewConnectionArguments(p.Args)
			return graphqlutil.NewConnectionFromArray(scopeIfaces, args), nil
		},
	})
}

var connResource = graphqlutil.NewConnectionDef(nodeResource)
