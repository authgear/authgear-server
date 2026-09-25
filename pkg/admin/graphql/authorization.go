package graphql

import (
	"context"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeAuthorization = "Authorization"
const typeAuthorizedScope = "AuthorizedScope"

var authorizedScopeType = graphql.NewObject(graphql.ObjectConfig{
	Name: typeAuthorizedScope,
	Fields: graphql.Fields{
		"resource": &graphql.Field{
			Type:        nodeResource,
			Description: "The Resource the scope is granted on. Null for a project-level scope.",
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(model.AuthorizedScope)
				if source.ResourceID == "" {
					return nil, nil
				}
				gqlCtx := GQLContext(p.Context)
				return gqlCtx.Resources.Load(p.Context, source.ResourceID).Value, nil
			},
		},
		"scope": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"description": &graphql.Field{
			Type:        graphql.String,
			Description: "The Scope's description. Null for a project-level scope.",
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(model.AuthorizedScope)
				if source.ResourceID == "" {
					return nil, nil
				}
				gqlCtx := GQLContext(p.Context)
				key := loader.ResourceScopeKey{ResourceID: source.ResourceID, Scope: source.Scope}
				return gqlCtx.ResourceScopes.Load(p.Context, key).Map(func(value any) (any, error) {
					if scope, _ := value.(*model.Scope); scope != nil {
						return scope.Description, nil
					}
					return nil, nil
				}).Value, nil
			},
		},
	},
})

// resolveAuthorizedScopes drops scopes granted on a resource that no longer
// exists. It returns a thunk so that the loads batch across authorizations.
func resolveAuthorizedScopes(ctx context.Context, gqlCtx *Context, authz *model.Authorization) func() (any, error) {
	resources := make([]*graphqlutil.Lazy, len(authz.AuthorizedScopes))
	for i, s := range authz.AuthorizedScopes {
		if s.ResourceID != "" {
			resources[i] = gqlCtx.Resources.Load(ctx, s.ResourceID)
		}
	}

	return func() (any, error) {
		out := []model.AuthorizedScope{}
		for i, s := range authz.AuthorizedScopes {
			if resources[i] != nil {
				resource, err := resources[i].Value()
				if err != nil {
					return nil, err
				}
				if r, _ := resource.(*model.Resource); r == nil {
					continue
				}
			}
			out = append(out, s)
		}
		return out, nil
	}
}

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
				Type:              graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
				Description:       "The granted project-level scopes. Excludes Resource Scopes.",
				DeprecationReason: "Use authorizedScopes.",
			},
			"authorizedScopes": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(authorizedScopeType))),
				Description: "The granted scopes. Excludes scopes granted on a Resource that has been deleted.",
				Resolve: func(p graphql.ResolveParams) (any, error) {
					ctx := p.Context
					gqlCtx := GQLContext(ctx)
					return resolveAuthorizedScopes(ctx, gqlCtx, p.Source.(*model.Authorization)), nil
				},
			},
		},
	}),
	&model.Authorization{},
	func(ctx context.Context, gqlCtx *Context, id string) (any, error) {
		authz, err := gqlCtx.AuthorizationFacade.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		return authz.ToAPIModel(), nil
	},
)

var connAuthorization = graphqlutil.NewConnectionDef(nodeAuthorization)
