package graphql

import (
	"github.com/graphql-go/graphql"
)

var collaborator = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Collaborator",
	Description: "Collaborator of an app",
	Fields: graphql.Fields{
		"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"userID":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
	},
})
