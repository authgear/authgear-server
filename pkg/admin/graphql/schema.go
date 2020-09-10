package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var Schema *graphql.Schema

func init() {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:      query,
		Extensions: []graphql.Extension{&graphqlutil.APIErrorExtension{}},
	})
	if err != nil {
		panic(err)
	}
	Schema = &schema
}
