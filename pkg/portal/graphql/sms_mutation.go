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
			Type:        smsProviderConfigurationTwilioInput,
			Description: "Configuration of Twilio",
		},
		"webhook": &graphql.InputObjectFieldConfig{
			Type:        smsProviderConfigurationWebhookInput,
			Description: "Configuration of Webhook",
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
			Type: graphql.NewNonNull(graphql.String),
		},
		"messagingServiceSID": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var smsProviderConfigurationWebhookInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SMSProviderConfigurationWebhookInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"url": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"timeout": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
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

			if providerConfig.Twilio != nil {
				err = gqlCtx.SMSService.SendByTwilio(ctx, app, to, *providerConfig.Twilio)
				if err != nil {
					return nil, err
				}
			} else if providerConfig.Webhook != nil {
				err = gqlCtx.SMSService.SendByWebhook(ctx, app, to, *providerConfig.Webhook)
				if err != nil {
					return nil, err
				}
			}

			return nil, nil
		},
	},
)
