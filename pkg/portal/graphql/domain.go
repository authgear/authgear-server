package graphql

import (
	"github.com/graphql-go/graphql"
)

var domain = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Domain",
	Description: "DNS domain of an app",
	Fields: graphql.Fields{
		"id":                    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"createdAt":             &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
		"domain":                &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"apexDomain":            &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"verificationDNSRecord": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"isVerified":            &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
	},
})
