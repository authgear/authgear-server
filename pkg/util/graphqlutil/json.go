package graphqlutil

import (
	"encoding/json"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

var JSONObject = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "JSONObject",
	Description: "The `JSONObject` scalar type represents an arbitrary JSON object",
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
