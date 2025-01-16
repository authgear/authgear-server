package graphql

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/model"
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"
)

var sendTestSMSInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SendTestSMSInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID to test.",
		},
		"to": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The recipient phone number.",
		},
		"providerConfiguration": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(smsProviderConfigurationInput),
			Description: "The sms provider configuration.",
		},
	},
})

var smsProviderConfigurationInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"twilio": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Configuration of Twilio",
		},
	},
})

var smsProviderConfigurationTwilioInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationTwilioInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"accountSID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"authToken": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"messagingServiceSID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var _ = registerMutationField(
	"sendTestSMTPConfigurationEmail",
	&graphql.Field{
		Description: "Send test STMP configuration email",
		Type:        graphql.Boolean,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(sendTestSMTPConfigurationEmailInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			to := input["to"].(string)
			providerConfigurationInput := input["providerConfiguration"]

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			// Access control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
			if err != nil {
				return nil, err
			}

			app, err := gqlCtx.AppService.Get(ctx, appID)
			if err != nil {
				return nil, err
			}

			providerConfigJSON, err := json.Marshal(providerConfigurationInput)
			if err != nil {
				return nil, err
			}
			var providerConfig model.SMSProviderConfigurationInput
			err = json.Unmarshal(providerConfigJSON, &providerConfig)
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	},
)
