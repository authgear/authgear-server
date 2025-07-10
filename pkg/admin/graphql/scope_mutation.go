package graphql

import (
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/graphql-go/graphql"
)

var createScopeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateScopeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the resource.",
		},
		"scope": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The scope string.",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The optional description of the scope.",
		},
	},
})

var createScopePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateScopePayload",
	Fields: graphql.Fields{
		"scope": &graphql.Field{
			Type: graphql.NewNonNull(nodeScope),
		},
	},
})

var _ = registerMutationField(
	"createScope",
	&graphql.Field{
		Description: "Create a new scope.",
		Type:        graphql.NewNonNull(createScopePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createScopeInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			resourceID := input["resourceID"].(string)
			resourceNodeID := relay.FromGlobalID(resourceID)
			if resourceNodeID == nil || resourceNodeID.Type != typeResource {
				return nil, ErrInvalidResourceID
			}
			decodedResourceID := resourceNodeID.ID
			scopeStr := input["scope"].(string)

			var description *string
			if str, ok := input["description"].(string); ok && str != "" {
				description = &str
			}

			options := &resourcescope.NewScopeOptions{
				ResourceID:  decodedResourceID,
				Scope:       scopeStr,
				Description: description,
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			scope, err := gqlCtx.ResourceScopeFacade.CreateScope(ctx, options)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationCreateScopeExecutedEventPayload{
				Scope: *scope,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"scope": scope,
			}).Value, nil
		},
	},
)

var updateScopeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateScopeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the scope.",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new description of the scope. Pass null if you do not need to update the description. Pass an empty string to remove the description.",
		},
	},
})

var updateScopePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateScopePayload",
	Fields: graphql.Fields{
		"scope": &graphql.Field{
			Type: graphql.NewNonNull(nodeScope),
		},
	},
})

var _ = registerMutationField(
	"updateScope",
	&graphql.Field{
		Description: "Update an existing scope.",
		Type:        graphql.NewNonNull(updateScopePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateScopeInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			scopeNodeID := input["id"].(string)
			resolvedNodeID := relay.FromGlobalID(scopeNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeScope {
				return nil, ErrInvalidScopeID
			}
			scopeID := resolvedNodeID.ID

			var newDescription *string
			if str, ok := input["description"].(string); ok {
				newDescription = &str
			}

			options := &resourcescope.UpdateScopeOptions{
				ID:      scopeID,
				NewDesc: newDescription,
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			scope, err := gqlCtx.ResourceScopeFacade.UpdateScope(ctx, options)
			if err != nil {
				return nil, err
			}

			originalScope, err := gqlCtx.ResourceScopeFacade.GetScope(ctx, scopeID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationUpdateScopeExecutedEventPayload{
				OriginalScope: *originalScope,
				NewScope:      *scope,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"scope": scope,
			}).Value, nil
		},
	},
)

var deleteScopeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteScopeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the scope.",
		},
	},
})

var deleteScopePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteScopePayload",
	Fields: graphql.Fields{
		"ok": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var _ = registerMutationField(
	"deleteScope",
	&graphql.Field{
		Description: "Delete a scope.",
		Type:        graphql.NewNonNull(deleteScopePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteScopeInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			scopeNodeID := input["id"].(string)
			resolvedNodeID := relay.FromGlobalID(scopeNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeScope {
				return nil, ErrInvalidScopeID
			}
			scopeID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			err := gqlCtx.ResourceScopeFacade.DeleteScope(ctx, scopeID)
			if err != nil {
				return nil, err
			}

			scope, err := gqlCtx.ResourceScopeFacade.GetScope(ctx, scopeID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationDeleteScopeExecutedEventPayload{
				Scope: *scope,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"ok": true,
			}, nil
		},
	},
)
