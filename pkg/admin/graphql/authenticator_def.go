package graphql

import (
	"github.com/graphql-go/graphql"
)

var authenticatorDefOOBOTPEmail = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AuthenticatorDefinitionOOBOTPEmail",
	Fields: graphql.InputObjectConfigFieldMap{
		"email": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Email of the new oob otp sms authenticator.",
		},
	},
})

var authenticatorDefOOBOTPSMS = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AuthenticatorDefinitionOOBOTPSMS",
	Fields: graphql.InputObjectConfigFieldMap{
		"phone": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Phone number of the new oob otp sms authenticator.",
		},
	},
})

var authenticatorDefOOBOTPPassword = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AuthenticatorDefinitionPassword",
	Fields: graphql.InputObjectConfigFieldMap{
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Password of the new authenticator.",
		},
	},
})

var authenticatorDef = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "AuthenticatorDefinition",
	Description: "Definition of an authenticator. This is a union object, exactly one of the available fields must be present.",
	Fields: graphql.InputObjectConfigFieldMap{
		"kind": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(authenticatorKind),
			Description: "Kind of authenticator.",
		},
		"type": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(authenticatorType),
			Description: "Type of authenticator.",
		},
		"oobOtpEmail": &graphql.InputObjectFieldConfig{
			Type:        authenticatorDefOOBOTPEmail,
			Description: "OOB OTP Email authenticator definition. Must be provided when type is OOB_OTP_EMAIL.",
		},
		"oobOtpSMS": &graphql.InputObjectFieldConfig{
			Type:        authenticatorDefOOBOTPSMS,
			Description: "OOB OTP SMS authenticator definition. Must be provided when type is OOB_OTP_SMS.",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        authenticatorDefOOBOTPPassword,
			Description: "Password authenticator definition. Must be provided when type is PASSWORD.",
		},
	},
})
