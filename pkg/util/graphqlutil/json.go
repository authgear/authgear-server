package graphqlutil

import (
	"encoding/json"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func NewJSONObjectScalar(name string, description string) *graphql.Scalar {
	return graphql.NewScalar(graphql.ScalarConfig{
		Name:        name,
		Description: description,
		Serialize: func(value interface{}) interface{} {
			return value
		},
		ParseValue: func(value interface{}) interface{} {
			return value
		},
		ParseLiteral: func(valueAST ast.Value) interface{} {
			switch valueAST := valueAST.(type) {
			case *ast.StringValue:
				var obj interface{}
				if err := json.Unmarshal([]byte(valueAST.Value), &obj); err == nil {
					return obj
				}
			}
			return nil
		},
	})

}
