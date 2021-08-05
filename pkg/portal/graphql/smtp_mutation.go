package graphql

import (
	"github.com/graphql-go/graphql"
)

var sendTestSMTPConfigurationEmailInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "sendTestSMTPConfigurationEmailInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID to test.",
		},
		"to": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The recipient email address.",
		},
		"smtpHost": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "SMTP Host.",
		},
		"smtpPort": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "SMTP Port.",
		},
		"smtpUsername": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "SMTP Username.",
		},
		"smtpPassword": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "SMTP Password.",
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
			// TODO
			return nil, nil
		},
	},
)
