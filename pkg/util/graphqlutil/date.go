package graphqlutil

import (
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func serializeDate(value interface{}) interface{} {
	switch value := value.(type) {
	case time.Time:
		return value.Format("2006-01-02")
	case *time.Time:
		if value == nil {
			return nil
		}
		return serializeDate(*value)
	default:
		return nil
	}
}

func unserializeDate(value interface{}) interface{} {
	switch value := value.(type) {
	case []byte:
		return unserializeDate(string(value))
	case string:
		t, err := time.Parse("2006-01-02", value)
		if err != nil {
			return nil
		}
		return t
	case *string:
		if value == nil {
			return nil
		}
		return unserializeDate(*value)
	case time.Time:
		// Truncate time
		t := value.UTC()
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	case *time.Time:
		if value == nil {
			return nil
		}
		return unserializeDate(*value)
	default:
		return nil
	}
}

var Date = graphql.NewScalar(graphql.ScalarConfig{
	Name: "Date",
	Description: "The `Date` scalar type represents a Date." +
		" The Date is serialized in ISO 8601 format",
	Serialize:  serializeDate,
	ParseValue: unserializeDate,
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			return unserializeDate(valueAST.Value)
		}
		return nil
	},
})
