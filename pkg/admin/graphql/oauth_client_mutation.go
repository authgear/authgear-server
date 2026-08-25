package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
)

var deleteDynamicClientInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteDynamicClientInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"clientID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The client_id of the DCR-registered or CIMD-resolved client to delete.",
		},
	},
})

var deleteDynamicClientPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteDynamicClientPayload",
	Fields: graphql.Fields{
		"ok": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var _ = registerMutationField(
	"deleteDynamicClient",
	&graphql.Field{
		Description: "Deletes a DCR-registered or CIMD-resolved client and frees one slot against its client limit.",
		Type:        graphql.NewNonNull(deleteDynamicClientPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteDynamicClientInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			input := p.Args["input"].(map[string]any)
			clientID := input["clientID"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			client, err := gqlCtx.DCRFacade.DeleteClient(ctx, clientID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationDeleteDynamicClientExecutedEventPayload{
				ClientID:   client.ClientID,
				Source:     client.Source,
				Kind:       client.Kind,
				ClientName: client.Name,
			})
			if err != nil {
				return nil, err
			}

			return map[string]any{"ok": true}, nil
		},
	},
)
