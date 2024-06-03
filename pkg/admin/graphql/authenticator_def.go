package graphql

import (
	"github.com/graphql-go/graphql"
)

var authenticatorDefOOBOTPEmail = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AuthenticatorDefinitionOOBOTPEmail",
	Fields: graphql.InputObjectConfigFieldMap{
		"email": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Email of the new oob otp sms authenticator.",
		},
	},
})

var authenticatorDefOOBOTPSMS = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AuthenticatorDefinitionOOBOTPSMS",
	Fields: graphql.InputObjectConfigFieldMap{
		"phone": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Phone number of the new oob otp sms authenticator.",
		},
	},
})

var authenticatorDefOOBOTPPassword = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AuthenticatorDefinitionPassword",
	Fields: graphql.InputObjectConfigFieldMap{
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Password of the new authenticator.",
		},
	},
})

var authenticatorDef = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "IdentityDefinition",
	Description: "Definition of an identity. This is a union object, exactly one of the available fields must be present.",
	Fields: graphql.InputObjectConfigFieldMap{
		"kind": &graphql.InputObjectFieldConfig{
			Type:        authenticatorKind,
			Description: "Kind of authenticator",
		},
		"oobOtpEmail": &graphql.InputObjectFieldConfig{
			Type:        authenticatorDefOOBOTPEmail,
			Description: "OOB OTP Email authenticator definition.",
		},
		"oobOtpSMS": &graphql.InputObjectFieldConfig{
			Type:        authenticatorDefOOBOTPSMS,
			Description: "OOB OTP SMS authenticator definition.",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        authenticatorDefOOBOTPPassword,
			Description: "Password authenticator definition.",
		},
	},
})
