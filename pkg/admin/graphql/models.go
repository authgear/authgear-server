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

var userAgent = graphql.NewObject(graphql.ObjectConfig{
	Name: "UserAgent",
	Fields: graphql.Fields{
		"raw":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"name":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"version":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"os":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"osVersion":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"deviceModel": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})
