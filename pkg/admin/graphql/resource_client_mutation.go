package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var addResourceToClientIDInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddResourceToClientIDInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceURI": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"clientID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The client ID to associate.",
		},
	},
})

var addResourceToClientIDPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddResourceToClientIDPayload",
	Fields: graphql.Fields{
		"resource": &graphql.Field{
			Type: graphql.NewNonNull(nodeResource),
		},
	},
})

var _ = registerMutationField(
	"addResourceToClientID",
	&graphql.Field{
		Description: "Associate a resource with a clientID.",
		Type:        graphql.NewNonNull(addResourceToClientIDPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addResourceToClientIDInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			resourceURI := input["resourceURI"].(string)
			clientID := input["clientID"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.ResourceScopeFacade.AddResourceToClientID(ctx, resourceURI, clientID)
			if err != nil {
				return nil, err
			}
			resource, err := gqlCtx.ResourceScopeFacade.GetResourceByURI(ctx, resourceURI)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationAddResourceToClientIDExecutedEventPayload{
				Resource: *resource,
				ClientID: clientID,
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

var removeResourceFromClientIDInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveResourceFromClientIDInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceURI": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"clientID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The client ID to disassociate.",
		},
	},
})

var removeResourceFromClientIDPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveResourceFromClientIDPayload",
	Fields: graphql.Fields{
		"resource": &graphql.Field{
			Type: graphql.NewNonNull(nodeResource),
		},
	},
})

var _ = registerMutationField(
	"removeResourceFromClientID",
	&graphql.Field{
		Description: "Disassociate a resource from a clientID.",
		Type:        graphql.NewNonNull(removeResourceFromClientIDPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeResourceFromClientIDInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			resourceURI := input["resourceURI"].(string)
			clientID := input["clientID"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.ResourceScopeFacade.RemoveResourceFromClientID(ctx, resourceURI, clientID)
			if err != nil {
				return nil, err
			}
			resource, err := gqlCtx.ResourceScopeFacade.GetResourceByURI(ctx, resourceURI)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload{
				Resource: *resource,
				ClientID: clientID,
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
