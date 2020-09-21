package graphql

import "github.com/graphql-go/graphql"

var claim = graphql.NewObject(graphql.ObjectConfig{
	Name: "Claim",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"value": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})
