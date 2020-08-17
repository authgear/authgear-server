package graphql

import "github.com/graphql-go/graphql"

var Schema *graphql.Schema

func init() {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: query,
	})
	if err != nil {
		panic(err)
	}
	Schema = &schema
}
