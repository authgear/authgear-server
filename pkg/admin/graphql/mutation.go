package graphql

import (
	"github.com/graphql-go/graphql"
)

func registerMutationField(name string, field *graphql.Field) *graphql.Field {
	mutationFields[name] = field
	return field
}

var mutationFields = graphql.Fields{}

var mutation = graphql.NewObject(graphql.ObjectConfig{
	Name:   "Mutation",
	Fields: mutationFields,
})
