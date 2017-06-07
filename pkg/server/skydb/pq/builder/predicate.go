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
	"bytes"
	"encoding/json"
	"fmt"

	sq "github.com/lann/squirrel"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

// accessPredicateSqlizer build the json matching expression base on user's
// role, the builded express will filter out record which user is not accessible.
//
// The sql for record accessible by user rickmak
// `_access @> '[{"user_id":"rickmak"}]'`
//
// Record accessible by user with admin role
// `_access @> '[{"role":"admin"}]'`
//
// Record accessible by user rickmak or admin role
// `_access @> '[{"role":"rickmak"}]' OR _access @> '[{"role":"admin"}]'`¬
type accessPredicateSqlizer struct {
	user  *skydb.UserInfo
	level skydb.RecordACLLevel
}

func (p accessPredicateSqlizer) ToSql() (string, []interface{}, error) {
	var b bytes.Buffer
	b.WriteString(`(`)
	args := []interface{}{}

	if p.user != nil {
		if p.user.ID == "" {
			panic("cannot build access predicate without user")
		}

		escapedID, err := json.Marshal(p.user.ID)
		if err != nil {
			panic("unexpected serialize error on user_id")
		}

		for _, role := range p.user.Roles {
			escapedRole, err := json.Marshal(role)
			if err != nil {
				panic("unexpected serialize error on role")
			}
			b.WriteString(fmt.Sprintf(`_access @> '[{"role": %s}]' OR `, escapedRole))
		}
		b.WriteString(fmt.Sprintf(`_access @> '[{"user_id": %s}]' OR `, escapedID))

		b.WriteString(`_owner_id = ? OR `)
		args = append(args, p.user.ID)
	}

	if p.level == skydb.ReadLevel {
		b.WriteString(`_access @> '[{"public": true}]' OR `)
	} else if p.level == skydb.WriteLevel {
		b.WriteString(`_access @> '[{"public": true, "level": "write"}]' OR `)
	}

	b.WriteString(`_access IS NULL)`)

	return b.String(), args, nil
}

type userRelationPredicateSqlizer struct {
	outwardAlias string
	inwardAlias  string
	user         string
}

func (p userRelationPredicateSqlizer) ToSql() (sql string, args []interface{}, err error) {
	if p.outwardAlias != "" && p.inwardAlias != "" {
		sql = fmt.Sprintf("%s = %s AND %s = ?",
			fullQuoteIdentifier(p.outwardAlias, "left_id"),
			fullQuoteIdentifier(p.inwardAlias, "right_id"),
			fullQuoteIdentifier(p.outwardAlias, "left_id"))
	} else if p.outwardAlias != "" {
		sql = fmt.Sprintf("%s = ?",
			fullQuoteIdentifier(p.outwardAlias, "left_id"))
	} else if p.inwardAlias != "" {
		sql = fmt.Sprintf("%s = ?",
			fullQuoteIdentifier(p.inwardAlias, "right_id"))
	} else {
		panic("unexpected value in sqlizer")
	}
	args = []interface{}{p.user}
	err = nil
	return
}

type containsComparisonPredicateSqlizer struct {
	sqlizers []expressionSqlizer
}

func (p *containsComparisonPredicateSqlizer) ToSql() (sql string, args []interface{}, err error) {
	var buffer bytes.Buffer
	lhs := p.sqlizers[0]
	rhs := p.sqlizers[1]

	if lhs.fieldType.Type.IsGeometryCompatibleType() && rhs.fieldType.Type.IsGeometryCompatibleType() {
		buffer.WriteString(`ST_Contains(`)

		sqlOperand, opArgs, err := rhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		buffer.WriteString(`, `)

		sqlOperand, opArgs, err = lhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		buffer.WriteString(`)`)

		sql = buffer.String()
		return sql, args, err
	} else if lhs.Type == skydb.Literal && rhs.Type == skydb.KeyPath {
		buffer.WriteString(`jsonb_exists(`)

		sqlOperand, opArgs, err := rhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		buffer.WriteString(`, `)

		sqlOperand, opArgs, err = lhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		buffer.WriteString(`)`)

		sql = buffer.String()
		return sql, args, err
	} else if lhs.Type == skydb.KeyPath && rhs.Type == skydb.Literal {
		sqlOperand, opArgs, err := lhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		buffer.WriteString(` IN `)

		sqlOperand, opArgs, err = rhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		sql = buffer.String()
		return sql, args, err
	}

	// Note: "In" operator may be used to compare other types of values
	// but the generated SQL depends on the types of values being compared.
	// It is currently not supported to compare two keypaths,
	// unless they are geometry types.  cf. #345
	return "", []interface{}{}, ErrCannotCompareUsingInOperator
}

type comparisonPredicateSqlizer struct {
	sqlizers []expressionSqlizer
	operator skydb.Operator
}

func (p *comparisonPredicateSqlizer) ToSql() (sql string, args []interface{}, err error) {
	args = []interface{}{}
	if p.operator.IsBinary() {
		var buffer bytes.Buffer
		lhs := p.sqlizers[0]
		rhs := p.sqlizers[1]

		if p.operator.IsCommutative() {
			if lhs.Expression.IsLiteralNull() && !rhs.Expression.IsLiteralNull() {
				// In SQL, NULL must be on the right side of a comparison
				// operator.
				lhs, rhs = rhs, lhs
			}
		}

		sqlOperand, opArgs, err := lhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		if rhs.IsLiteralNull() {
			err = p.writeOperatorForNullOperand(&buffer)
		} else {
			err = p.writeOperator(&buffer)
		}
		if err != nil {
			return "", nil, err
		}

		sqlOperand, opArgs, err = rhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		sql = buffer.String()
	} else {
		err = fmt.Errorf("comparison operator `%v` is not supported", p.operator)
	}

	return
}

func (p *comparisonPredicateSqlizer) writeOperator(buffer *bytes.Buffer) error {
	switch p.operator {
	default:
		return fmt.Errorf("comparison operator `%v` is not supported", p.operator)
	case skydb.Equal:
		buffer.WriteString(`=`)
	case skydb.GreaterThan:
		buffer.WriteString(`>`)
	case skydb.LessThan:
		buffer.WriteString(`<`)
	case skydb.GreaterThanOrEqual:
		buffer.WriteString(`>=`)
	case skydb.LessThanOrEqual:
		buffer.WriteString(`<=`)
	case skydb.NotEqual:
		buffer.WriteString(`<>`)
	case skydb.Like:
		buffer.WriteString(` LIKE `)
	case skydb.ILike:
		buffer.WriteString(` ILIKE `)
	}
	return nil
}

func (p *comparisonPredicateSqlizer) writeOperatorForNullOperand(buffer *bytes.Buffer) error {
	switch p.operator {
	default:
		return p.writeOperator(buffer)
	case skydb.Equal:
		buffer.WriteString(` IS `)
	case skydb.NotEqual:
		buffer.WriteString(` IS NOT `)
	}
	return nil
}

// NotSqlizer generates SQL condition that negates a boolean condition
type NotSqlizer struct {
	Predicate sq.Sqlizer
}

// ToSql generates SQL for NotSqlizer
func (s NotSqlizer) ToSql() (sql string, args []interface{}, err error) {
	sql, args, err = s.Predicate.ToSql()
	if err != nil {
		return
	}
	sql = fmt.Sprintf("NOT (%s)", sql)
	return
}

// FalseSqlizer generates SQL condition that evaluates to false
type FalseSqlizer struct {
}

// ToSql generates SQL for FalseSqlizer
func (s FalseSqlizer) ToSql() (sql string, args []interface{}, err error) {
	return "FALSE", []interface{}{}, nil
}

// distancePredicateSqlizer generates SQL condition that calculates if a
// location is within a certain distance.
type distancePredicateSqlizer struct {
	alias    string
	field    string
	location skydb.Location
	distance expressionSqlizer
}

// ToSql generates SQL for distancePredicateSqlizer
func (s distancePredicateSqlizer) ToSql() (sql string, args []interface{}, err error) {
	distanceSQL, distanceArgs, err := s.distance.ToSql()
	if err != nil {
		return
	}

	sql = fmt.Sprintf(
		"ST_DWithin(%s::geography, ST_MakePoint(?, ?)::geography, %s)",
		fullQuoteIdentifier(s.alias, s.field),
		distanceSQL,
	)
	args = []interface{}{s.location.Lng(), s.location.Lat()}
	args = append(args, distanceArgs...)
	return
}
