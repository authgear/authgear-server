package graphql

import (
	"context"
	"sort"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeAuthorization = "Authorization"
const typeAuthorizationScope = "AuthorizationScope"

const (
	authorizationScopeKindProject  = "PROJECT"
	authorizationScopeKindResource = "RESOURCE"
)

var authorizationScopeKind = graphql.NewEnum(graphql.EnumConfig{
	Name: "AuthorizationScopeKind",
	Values: graphql.EnumValueConfigMap{
		authorizationScopeKindProject: &graphql.EnumValueConfig{
			Value:       authorizationScopeKindProject,
			Description: "A project-level scope, such as openid, profile or email.",
		},
		authorizationScopeKindResource: &graphql.EnumValueConfig{
			Value:       authorizationScopeKindResource,
			Description: "A scope defined by a Resource.",
		},
	},
})

// authorizationScope is one entry of Authorization.resolvedScopes. Resource is
// nil for a project-level scope, and for a resource scope the project no
// longer defines.
type authorizationScope struct {
	Scope       string
	Kind        string
	Resource    *model.Resource
	Description *string
}

var authorizationScopeType = graphql.NewObject(graphql.ObjectConfig{
	Name:        typeAuthorizationScope,
	Description: "A granted scope name resolved against the project's configured Resources and Scopes. An authorization records scope names without the Resource they were granted for, so this resolution is by name.",
	Fields: graphql.Fields{
		"scope": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The granted scope name.",
			Resolve: func(p graphql.ResolveParams) (any, error) {
				return p.Source.(authorizationScope).Scope, nil
			},
		},
		"kind": &graphql.Field{
			Type:        graphql.NewNonNull(authorizationScopeKind),
			Description: "Whether the name is project-level or defined by a Resource.",
			Resolve: func(p graphql.ResolveParams) (any, error) {
				return p.Source.(authorizationScope).Kind, nil
			},
		},
		"description": &graphql.Field{
			Type:        graphql.String,
			Description: "The description configured on the Scope, when there is one.",
			Resolve: func(p graphql.ResolveParams) (any, error) {
				return p.Source.(authorizationScope).Description, nil
			},
		},
	},
})

func init() {
	// Registered here because nodeResource is defined in another file: at
	// package variable initialisation time its value is not ready yet.
	authorizationScopeType.AddFieldConfig("resource", &graphql.Field{
		Type:        nodeResource,
		Description: "The Resource defining this scope. Null for a project-level scope, and for a resource scope the project no longer defines.",
		Resolve: func(p graphql.ResolveParams) (any, error) {
			return p.Source.(authorizationScope).Resource, nil
		},
	})
}

// resolveScopes maps the granted scope names to their configured scopes. A
// name defined by more than one Resource yields one entry per Resource: the
// grant does not record which one it was granted for, and the same name on two
// Resources is two different permissions (see docs/specs/api-resource.md).
func resolveScopes(ctx context.Context, gqlCtx *Context, granted []string) ([]authorizationScope, error) {
	var resourceScopeNames []string
	for _, s := range granted {
		if oauth.IsResourceScope(s) {
			resourceScopeNames = append(resourceScopeNames, s)
		}
	}

	byName := make(map[string][]*model.Scope, len(resourceScopeNames))
	if len(resourceScopeNames) > 0 {
		configured, err := gqlCtx.ResourceScopeFacade.ListScopesByNames(ctx, resourceScopeNames)
		if err != nil {
			return nil, err
		}
		for _, sc := range configured {
			byName[sc.Scope] = append(byName[sc.Scope], sc)
		}
	}

	// Queue every resource load before forcing any of them, so the loader
	// batches them into one query instead of one per scope.
	lazyResources := make(map[string]*graphqlutil.Lazy, len(byName))
	for _, scopes := range byName {
		for _, sc := range scopes {
			if _, ok := lazyResources[sc.ResourceID]; !ok {
				lazyResources[sc.ResourceID] = gqlCtx.Resources.Load(ctx, sc.ResourceID)
			}
		}
	}
	resources := make(map[string]*model.Resource, len(lazyResources))
	for id, lazy := range lazyResources {
		value, err := lazy.Value()
		if err != nil {
			return nil, err
		}
		if r, ok := value.(*model.Resource); ok {
			resources[id] = r
		}
	}

	return buildAuthorizationScopes(granted, byName, resources), nil
}

// buildAuthorizationScopes is the pure part of resolveScopes: it decides each
// granted name's kind and attaches the resources that define it.
func buildAuthorizationScopes(
	granted []string,
	byName map[string][]*model.Scope,
	resources map[string]*model.Resource,
) []authorizationScope {
	var result []authorizationScope
	for _, name := range granted {
		if !oauth.IsResourceScope(name) {
			result = append(result, authorizationScope{Scope: name, Kind: authorizationScopeKindProject})
			continue
		}

		configured := byName[name]
		if len(configured) == 0 {
			// Granted before the scope was deleted from the project.
			result = append(result, authorizationScope{Scope: name, Kind: authorizationScopeKindResource})
			continue
		}

		// A name on two Resources arrives in an arbitrary store order; sort
		// so the same grant always renders the same way.
		entries := make([]authorizationScope, 0, len(configured))
		for _, sc := range configured {
			entries = append(entries, authorizationScope{
				Scope:       name,
				Kind:        authorizationScopeKindResource,
				Resource:    resources[sc.ResourceID],
				Description: sc.Description,
			})
		}
		sort.SliceStable(entries, func(i, j int) bool {
			return resourceSortKey(entries[i].Resource) < resourceSortKey(entries[j].Resource)
		})
		result = append(result, entries...)
	}

	return result
}

func resourceSortKey(r *model.Resource) string {
	if r == nil {
		return ""
	}
	return r.ResourceURI
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
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			},
			"resolvedScopes": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(authorizationScopeType))),
				Description: "The granted scopes, each resolved against the project's configured Resources and Scopes. One entry per granted scope name, except that a name defined by more than one Resource yields one entry per Resource.",
				Resolve: func(p graphql.ResolveParams) (any, error) {
					authz := p.Source.(*model.Authorization)
					ctx := p.Context
					gqlCtx := GQLContext(ctx)

					return resolveScopes(ctx, gqlCtx, authz.Scopes)
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
