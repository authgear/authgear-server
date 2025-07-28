package graphql

import (
	"context"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
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
			"resourceURI": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The URI of the resource.",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional name of the resource.",
			},
			"clientIDs": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
				Description: "The list of client IDs associated with this Resource.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					resource := p.Source.(*model.Resource)
					ctx := p.Context
					gqlCtx := GQLContext(ctx)

					clientIDs, err := gqlCtx.ResourceScopeFacade.ListClientIDsByResourceID(ctx, resource.ID)
					if err != nil {
						return nil, err
					}
					return clientIDs, nil
				},
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

			clientID, _ := p.Args["clientID"].(string)
			searchKeyword, _ := p.Args["searchKeyword"].(string)
			pageArgs := graphqlutil.NewPageArgsWithMaxPageSize(relay.NewConnectionArguments(p.Args), 1000)
			options := &resourcescope.ListScopeOptions{
				SearchKeyword: searchKeyword,
				ClientID:      clientID,
			}
			refs, result, err := gqlCtx.ResourceScopeFacade.ListScopes(ctx, source.ID, options, pageArgs)
			if err != nil {
				return nil, err
			}
			var lazyItems []graphqlutil.LazyItem
			for _, ref := range refs {
				lazyItems = append(lazyItems, graphqlutil.LazyItem{
					Lazy:   gqlCtx.Scopes.Load(ctx, ref.ID),
					Cursor: graphqlutil.Cursor(ref.Cursor),
				})
			}
			return graphqlutil.NewConnectionFromResult(lazyItems, result)
		},
	})
}

var connResource = graphqlutil.NewConnectionDef(nodeResource)
