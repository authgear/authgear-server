// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package builder

import (
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/lann/squirrel"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

// expressionSqlizer generates an SQL expression from a record.Expression. A SQL
// expression are those found in SELECT clause or in the WHERE clause.
//
// In addition to generating literal value such as string (`"hello"`) or integer (`1`),
// the expressionSqlizer can also generate expression for a column or a function.
//
// When expression is a column of a table, the `alias` field is required
// and it is either the name of the table of the column, or a SQL alias of such
// table.
type expressionSqlizer struct {
	// Alias is the name to qualify a SQL identifier. Could be empty
	// if there is no identifier or if the identifier does not need
	// to be qualified. This is usually the table name of the table alias.
	alias string

	// RequireCast is true when the expression should be casted to a different
	// SQL type. Whether the expression requires casting depends on the context
	// where the expression is used, hence cannot be determined by the
	// expression itself.
	requireCast bool

	// FieldType contains the database field type when available. If the
	// expression is a literal or is a computed value (such as function),
	// the field type maybe derived from the expression value. If not
	// available, the field type may be empty.
	fieldType record.FieldType

	record.Expression
}

func NewExpressionSqlizer(alias string, fieldType record.FieldType, expr record.Expression) sq.Sqlizer {
	return newExpressionSqlizer(alias, fieldType, expr)
}

func newExpressionSqlizer(alias string, fieldType record.FieldType, expr record.Expression) expressionSqlizer {
	requireCast := false
	if fieldType.Type.IsGeometryCompatibleType() && expr.Type == record.Literal {
		requireCast = true
	}

	return expressionSqlizer{
		alias,
		requireCast,
		fieldType,
		expr,
	}
}

func (expr expressionSqlizer) ToSql() (sql string, args []interface{}, err error) {
	switch expr.Type {
	case record.KeyPath:
		components := expr.KeyPathComponents()
		lastComponent := components[len(components)-1]
		sql = fullQuoteIdentifier(expr.alias, lastComponent)
		args = []interface{}{}

		if expr.requireCast {
			switch expr.fieldType.Type {
			case record.TypeLocation, record.TypeGeometry:
				sql = fmt.Sprintf("ST_AsGeoJSON(%s)", sql)
			}
		}
	case record.Function:
		sql, args = funcToSQLOperand(expr.alias, expr.Value.(record.Func))
	default:
		sql, args = LiteralToSQLOperand(expr.Value)
	}
	return
}

func RequireCast(sqlizer sq.Sqlizer) (sq.Sqlizer, error) {
	expr, ok := sqlizer.(expressionSqlizer)
	if !ok {
		return nil, errors.New("sqlizer not supported")
	}

	expr.requireCast = true
	return expr, nil
}

func funcToSQLOperand(alias string, fun record.Func) (string, []interface{}) {
	switch f := fun.(type) {
	case record.DistanceFunc:
		sql := fmt.Sprintf("ST_Distance_Sphere(%s, ST_MakePoint(?, ?))",
			fullQuoteIdentifier(alias, f.Field))
		args := []interface{}{f.Location.Lng(), f.Location.Lat()}
		return sql, args
	case record.CountFunc:
		var sql string
		if f.OverallRecords {
			sql = fmt.Sprintf("COUNT(*) OVER()")
		} else {
			sql = fmt.Sprintf("COUNT(*)")
		}
		args := []interface{}{}
		return sql, args
	default:
		panic(fmt.Errorf("got unrecgonized record.Func = %T", fun))
	}
}

func LiteralToSQLOperand(literal interface{}) (string, []interface{}) {
	// Array detection is borrowed from squirrel's expr.go
	switch literalValue := literal.(type) {
	case record.Geometry:
		valueInJSON, err := json.Marshal(literalValue)
		if err != nil {
			panic(fmt.Sprintf("unable to marshal record.Geometry: %s", err))
		}
		return fmt.Sprintf("ST_GeomFromGeoJSON(%s)", sq.Placeholders(1)), []interface{}{valueInJSON}
	case record.Location:
		return fmt.Sprintf("ST_MakePoint(%s)", sq.Placeholders(2)), []interface{}{literalValue.Lng(), literalValue.Lat()}
	case []interface{}:
		argCount := len(literalValue)
		if argCount > 0 {
			args := make([]interface{}, len(literalValue))
			for i, val := range literalValue {
				args[i] = literalToSQLValue(val)
			}
			return "(" + sq.Placeholders(len(literalValue)) + ")", args
		}

		// NOTE(limouren): trick to make `field IN (...)` work for empty list
		// NULL field won't match the condition since NULL == NULL is falsy,
		// which renders `field IN(NULL)` equivalent to FALSE
		return "(NULL)", nil
	default:
		if literal == nil {
			return "NULL", []interface{}{}
		}
		return sq.Placeholders(1), []interface{}{literalToSQLValue(literal)}
	}
}

func literalToSQLValue(value interface{}) interface{} {
	switch v := value.(type) {
	case record.Reference:
		return v.ID.Key
	default:
		return value
	}
}
