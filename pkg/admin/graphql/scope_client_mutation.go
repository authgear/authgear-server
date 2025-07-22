package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
)

var addScopesToClientIDInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddScopesToClientIDInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceURI": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"scopes": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "The list of scopes to add.",
		},
		"clientID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The client ID.",
		},
	},
})

var addScopesToClientIDPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddScopesToClientIDPayload",
	Fields: graphql.Fields{
		"scopes": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeScope))),
		},
	},
})

var _ = registerMutationField(
	"addScopesToClientID",
	&graphql.Field{
		Description: "Associate multiple scopes with a clientID.",
		Type:        graphql.NewNonNull(addScopesToClientIDPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addScopesToClientIDInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			resourceURI := input["resourceURI"].(string)
			scopesIface := input["scopes"].([]interface{})
			clientID := input["clientID"].(string)

			scopes := make([]string, len(scopesIface))
			for i, s := range scopesIface {
				scopes[i] = s.(string)
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			finalscopes, err := gqlCtx.ResourceScopeFacade.AddScopesToClientID(ctx, resourceURI, clientID, scopes)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationAddScopesToClientIDExecutedEventPayload{
				ResourceURI: resourceURI,
				ClientID:    clientID,
				Scopes:      finalscopes,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"scopes": finalscopes,
			}, nil
		},
	},
)

var removeScopesFromClientIDInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveScopesFromClientIDInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceURI": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"scopes": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "The list of scopes to remove.",
		},
		"clientID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The client ID.",
		},
	},
})

var removeScopesFromClientIDPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveScopesFromClientIDPayload",
	Fields: graphql.Fields{
		"scopes": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeScope))),
		},
	},
})

var _ = registerMutationField(
	"removeScopesFromClientID",
	&graphql.Field{
		Description: "Disassociate multiple scopes from a clientID.",
		Type:        graphql.NewNonNull(removeScopesFromClientIDPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeScopesFromClientIDInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			resourceURI := input["resourceURI"].(string)
			scopesIface := input["scopes"].([]interface{})
			clientID := input["clientID"].(string)

			scopes := make([]string, len(scopesIface))
			for i, s := range scopesIface {
				scopes[i] = s.(string)
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			finalscopes, err := gqlCtx.ResourceScopeFacade.RemoveScopesFromClientID(ctx, resourceURI, clientID, scopes)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload{
				ResourceURI: resourceURI,
				ClientID:    clientID,
				Scopes:      finalscopes,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"scopes": finalscopes,
			}, nil
		},
	},
)

var replaceScopesOfClientIDInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ReplaceScopesOfClientIDInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"resourceURI": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The URI of the resource.",
		},
		"clientID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The client ID.",
		},
		"scopes": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "The new list of scopes.",
		},
	},
})

var replaceScopesOfClientIDPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "ReplaceScopesOfClientIDPayload",
	Fields: graphql.Fields{
		"scopes": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeScope))),
		},
	},
})

var _ = registerMutationField(
	"replaceScopesOfClientID",
	&graphql.Field{
		Description: "Replace the set of scopes associated with a clientID.",
		Type:        graphql.NewNonNull(replaceScopesOfClientIDPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(replaceScopesOfClientIDInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			resourceURI := input["resourceURI"].(string)
			clientID := input["clientID"].(string)
			scopesIface := input["scopes"].([]interface{})
			scopes := make([]string, len(scopesIface))
			for i, s := range scopesIface {
				scopes[i] = s.(string)
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			finalscopes, err := gqlCtx.ResourceScopeFacade.ReplaceScopesOfClientID(ctx, resourceURI, clientID, scopes)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload{
				ResourceURI: resourceURI,
				ClientID:    clientID,
				Scopes:      finalscopes,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"scopes": finalscopes,
			}, nil
		},
	},
)
