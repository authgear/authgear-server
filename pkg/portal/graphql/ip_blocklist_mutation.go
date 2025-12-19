package graphql

import (
	"context"

	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/util/validation"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var checkIPInputSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"ipAddress": {
				"type": "string",
				"format": "x_ip"
			},
			"cidrs": {
				"type": "array",
				"items": {
					"type": "string",
					"format": "x_cidr"
				}
			},
			"countryCodes": {
				"type": "array",
				"items": {
					"type": "string",
					"minLength": 2,
					"maxLength": 2
				}
			}
		},
		"required": ["ipAddress", "cidrs", "countryCodes"]
	}
`)

var checkIPInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CheckIPInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID.",
		},
		"ipAddress": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The IP address to check.",
		},
		"cidrs": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "List of CIDRs to check against.",
		},
		"countryCodes": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "List of alpha-2 country codes to check against.",
		},
	},
})

var _ = registerMutationField(
	"checkIP",
	&graphql.Field{
		Description: "Check an IP address against blocklists.",
		Type:        graphql.Boolean,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(checkIPInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			err := checkIPInputSchema.Validator().ValidateValue(context.Background(), input)
			if err != nil {
				return nil, err
			}

			appNodeID := input["appID"].(string)
			ipAddress := input["ipAddress"].(string)
			cidrs := input["cidrs"].([]interface{})
			countryCodes := input["countryCodes"].([]interface{})

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			// Access control: collaborator.
			_, err = gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
			if err != nil {
				return nil, err
			}

			cidrsStr := make([]string, len(cidrs))
			for i, v := range cidrs {
				cidrsStr[i] = v.(string)
			}

			countryCodesStr := make([]string, len(countryCodes))
			for i, v := range countryCodes {
				countryCodesStr[i] = v.(string)
			}

			ok := gqlCtx.IPBlocklistService.CheckIP(ctx, ipAddress, cidrsStr, countryCodesStr)
			return ok, nil
		},
	},
)
