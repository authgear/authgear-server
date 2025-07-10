package graphql

import (
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/graphql-go/graphql"
)

var createResourceInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateResourceInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"uri": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"name": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The optional name of the resource.",
		},
	},
})

var createResourcePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateResourcePayload",
	Fields: graphql.Fields{
		"resource": &graphql.Field{
			Type: graphql.NewNonNull(nodeResource),
		},
	},
})

var _ = registerMutationField(
	"createResource",
	&graphql.Field{
		Description: "Create a new resource.",
		Type:        graphql.NewNonNull(createResourcePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createResourceInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			uri := input["uri"].(string)

			var name *string
			if str, ok := input["name"].(string); ok && str != "" {
				name = &str
			}

			options := &resourcescope.NewResourceOptions{
				URI:  uri,
				Name: name,
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			resource, err := gqlCtx.ResourceScopeFacade.CreateResource(ctx, options)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationCreateResourceExecutedEventPayload{
				Resource: *resource,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"resource": resource,
			}).Value, nil
		},
	},
)

var updateResourceInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateResourceInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the resource.",
		},
		"name": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new name of the resource. Pass null if you do not need to update the name. Pass an empty string to remove the name.",
		},
	},
})

var updateResourcePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateResourcePayload",
	Fields: graphql.Fields{
		"resource": &graphql.Field{
			Type: graphql.NewNonNull(nodeResource),
		},
	},
})

var _ = registerMutationField(
	"updateResource",
	&graphql.Field{
		Description: "Update an existing resource.",
		Type:        graphql.NewNonNull(updateResourcePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateResourceInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			resourceNodeID := input["id"].(string)
			resolvedNodeID := relay.FromGlobalID(resourceNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeResource {
				return nil, ErrInvalidResourceID
			}
			resourceID := resolvedNodeID.ID

			var newName *string
			if str, ok := input["name"].(string); ok {
				newName = &str
			}

			options := &resourcescope.UpdateResourceOptions{
				ID:      resourceID,
				NewName: newName,
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			resource, err := gqlCtx.ResourceScopeFacade.UpdateResource(ctx, options)
			if err != nil {
				return nil, err
			}

			originalResource, err := gqlCtx.ResourceScopeFacade.GetResource(ctx, resourceID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationUpdateResourceExecutedEventPayload{
				OriginalResource: *originalResource,
				NewResource:      *resource,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"resource": resource,
			}).Value, nil
		},
	},
)

var deleteResourceInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteResourceInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the resource.",
		},
	},
})

var deleteResourcePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteResourcePayload",
	Fields: graphql.Fields{
		"ok": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var _ = registerMutationField(
	"deleteResource",
	&graphql.Field{
		Description: "Delete a resource.",
		Type:        graphql.NewNonNull(deleteResourcePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteResourceInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			resourceNodeID := input["id"].(string)
			resolvedNodeID := relay.FromGlobalID(resourceNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeResource {
				return nil, ErrInvalidResourceID
			}
			resourceID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			err := gqlCtx.ResourceScopeFacade.DeleteResource(ctx, resourceID)
			if err != nil {
				return nil, err
			}

			resource, err := gqlCtx.ResourceScopeFacade.GetResource(ctx, resourceID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationDeleteResourceExecutedEventPayload{
				Resource: *resource,
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
