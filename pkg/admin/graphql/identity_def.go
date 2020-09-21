package graphql

import (
	"github.com/graphql-go/graphql"
)

var identityDefLoginID = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "IdentityDefinitionLoginID",
	Fields: graphql.InputObjectConfigFieldMap{
		"key": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The login ID key.",
		},
		"value": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The login ID.",
		},
	},
})

var identityDef = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "IdentityDefinition",
	Description: "Definition of an identity. This is a union object, exactly one of the available fields must be present.",
	Fields: graphql.InputObjectConfigFieldMap{
		"loginID": &graphql.InputObjectFieldConfig{
			Type:        identityDefLoginID,
			Description: "Login ID identity definition.",
		},
	},
})
