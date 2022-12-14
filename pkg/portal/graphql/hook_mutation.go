package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var checkDenoHookInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "sendDenoHookInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID.",
		},
		"content": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The content of the hook.",
		},
	},
})

var _ = registerMutationField(
	"checkDenoHook",
	&graphql.Field{
		Description: "Check Deno Hook",
		Type:        graphql.Boolean,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(checkDenoHookInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			content := input["content"].(string)

			gqlCtx := GQLContext(p.Context)

			// Access control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.DenoService.Check(p.Context, content)
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	},
)
