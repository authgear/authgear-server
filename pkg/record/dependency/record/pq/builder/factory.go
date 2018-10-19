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
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"

	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type PredicateSqlizerFactory interface {
	UpdateTypemap(typemap record.Schema) record.Schema
	AddJoinsToSelectBuilder(q sq.SelectBuilder) sq.SelectBuilder
	NewPredicateSqlizer(p record.Predicate) (sq.Sqlizer, error)
	NewAccessControlSqlizer(user *authinfo.AuthInfo, aclLevel record.ACLLevel) (sq.Sqlizer, error)
}

// predicateSqlizerFactory is a factory for creating sqlizer for predicate
type predicateSqlizerFactory struct {
	recordStore  record.Store
	sqlBuilder   db.SQLBuilder
	primaryTable string
	joinedTables []joinedTable
	extraColumns map[string]record.FieldType
}

func NewPredicateSqlizerFactory(recordStore record.Store, sqlBuilder db.SQLBuilder, primaryTable string) PredicateSqlizerFactory {
	return &predicateSqlizerFactory{
		recordStore:  recordStore,
		sqlBuilder:   sqlBuilder,
		primaryTable: primaryTable,
		joinedTables: []joinedTable{},
	}
}

func (f *predicateSqlizerFactory) NewPredicateSqlizer(p record.Predicate) (sq.Sqlizer, error) {
	if p.IsEmpty() {
		panic("no sqlizer can be created from an empty predicate")
	}

	if p.Operator == record.Functional {
		return f.newFunctionalPredicateSqlizer(p)
	}
	if p.Operator.IsCompound() {
		return f.newCompoundPredicateSqlizer(p)
	}
	return f.newComparisonPredicateSqlizer(p)
}

func (f *predicateSqlizerFactory) newCompoundPredicateSqlizer(p record.Predicate) (sq.Sqlizer, error) {
	switch p.Operator {
	default:
		err := fmt.Errorf("compound operator `%v` is not supported", p.Operator)
		return nil, err
	case record.And:
		and := make(sq.And, len(p.Children))
		for i, child := range p.Children {
			sqlizer, err := f.NewPredicateSqlizer(child.(record.Predicate))
			if err != nil {
				return nil, err
			}
			and[i] = sqlizer
		}
		return and, nil
	case record.Or:
		or := make(sq.Or, len(p.Children))
		for i, child := range p.Children {
			sqlizer, err := f.NewPredicateSqlizer(child.(record.Predicate))
			if err != nil {
				return nil, err
			}
			or[i] = sqlizer
		}
		return or, nil
	case record.Not:
		pred := p.Children[0].(record.Predicate)
		sqlizer, err := f.NewPredicateSqlizer(pred)
		if err != nil {
			return nil, err
		}
		return NotSqlizer{sqlizer}, nil
	}
}

func (f *predicateSqlizerFactory) newFunctionalPredicateSqlizer(predicate record.Predicate) (sq.Sqlizer, error) {
	expr := predicate.Children[0].(record.Expression)
	if expr.Type != record.Function {
		panic("unexpected expression in functional predicate")
	}
	switch fn := expr.Value.(type) {
	case record.UserRelationFunc:
		return f.newUserRelationFunctionalPredicateSqlizer(fn)
	default:
		panic("the specified function cannot be used as a functional predicate")
	}
}

func (f *predicateSqlizerFactory) newUserRelationFunctionalPredicateSqlizer(fn record.UserRelationFunc) (sq.Sqlizer, error) {
	table := fn.RelationName
	direction := fn.RelationDirection
	if direction == "" {
		direction = "outward"
	}
	primaryColumn := fn.KeyPath
	if primaryColumn == "_owner" || primaryColumn == "" {
		primaryColumn = "_owner_id"
	}

	var outwardAlias, inwardAlias string
	if direction == "outward" || direction == "mutual" {
		outwardAlias = f.createLeftJoin(table, primaryColumn, "right_id")
	}
	if direction == "inward" || direction == "mutual" {
		inwardAlias = f.createLeftJoin(table, primaryColumn, "left_id")
	}

	return userRelationPredicateSqlizer{
		outwardAlias: outwardAlias,
		inwardAlias:  inwardAlias,
		user:         fn.User,
	}, nil
}

func (f *predicateSqlizerFactory) NewAccessControlSqlizer(user *authinfo.AuthInfo, aclLevel record.ACLLevel) (sq.Sqlizer, error) {
	return &accessPredicateSqlizer{
		f.primaryTable,
		user,
		aclLevel,
	}, nil
}

func (f *predicateSqlizerFactory) newComparisonPredicateSqlizer(p record.Predicate) (sq.Sqlizer, error) {
	if sqlizer, ok := f.tryOptimizeDistancePredicate(p); ok {
		return sqlizer, nil
	}

	sqlizers := []expressionSqlizer{}
	for _, child := range p.Children {
		sqlizer, err := f.newExpressionSqlizer(child.(record.Expression))
		if err != nil {
			return nil, err
		}
		sqlizers = append(sqlizers, sqlizer)
	}

	if p.Operator == record.In {
		return &containsComparisonPredicateSqlizer{sqlizers}, nil
	}
	return &comparisonPredicateSqlizer{sqlizers, p.Operator}, nil
}

// tryOptimizeDistancePredicate returns a sqlizer that is more efficient
// at querying whether two points are within certain distance.
//
// If the predicate cannot be optimize or an error occurred generating
// an optimized sqlizer, the second value returned is false.
func (f *predicateSqlizerFactory) tryOptimizeDistancePredicate(p record.Predicate) (sq.Sqlizer, bool) {
	var tryFunc record.Expression
	var tryValue record.Expression
	if p.Operator == record.LessThan {
		tryFunc = p.Children[0].(record.Expression)
		tryValue = p.Children[1].(record.Expression)
	} else if p.Operator == record.GreaterThan {
		tryFunc = p.Children[1].(record.Expression)
		tryValue = p.Children[0].(record.Expression)
	} else {
		return nil, false
	}

	distanceFunc, ok := tryFunc.Value.(record.DistanceFunc)
	if !ok {
		return nil, false
	}

	distanceValue, err := f.newExpressionSqlizer(tryValue)
	if err != nil {
		return nil, false
	}

	return &distancePredicateSqlizer{
		f.primaryTable,
		distanceFunc.Field,
		distanceFunc.Location,
		distanceValue,
	}, true
}

func (f *predicateSqlizerFactory) newExpressionSqlizer(expr record.Expression) (expressionSqlizer, error) {
	if expr.IsKeyPath() {
		return f.newExpressionSqlizerForKeyPath(expr)
	}

	if expr.Type == record.Literal {
		var fieldType record.FieldType
		if expr.Value != nil {
			var err error
			fieldType, err = record.DeriveFieldType(expr.Value)
			if err != nil {
				return expressionSqlizer{}, err
			}
		}

		sqlizer := newExpressionSqlizer(f.primaryTable, fieldType, expr)
		return sqlizer, nil
	}

	if expr.Type == record.Function {
		funcInterface, ok := expr.Value.(record.Func)
		if !ok {
			panic(`expression value is not a function`)
		}
		return newExpressionSqlizer(f.primaryTable, record.FieldType{Type: funcInterface.DataType()}, expr), nil
	}

	return expressionSqlizer{}, skyerr.NewError(skyerr.RecordQueryInvalid,
		`unexpected expression type`)
}

func (f *predicateSqlizerFactory) newExpressionSqlizerForKeyPath(expr record.Expression) (expressionSqlizer, error) {
	if !expr.IsKeyPath() {
		panic("expression is not a key path")
	}

	components := expr.KeyPathComponents()
	keyPath := expr.Value.(string)
	if len(components) > 2 {
		return expressionSqlizer{}, skyerr.NewErrorf(skyerr.RecordQueryInvalid,
			`keypath "%s" with more than 2 components is not supported`, keyPath)
	}

	alias := f.primaryTable
	fields, err := record.TraverseColumnTypes(f.recordStore, f.primaryTable, keyPath)
	if err != nil {
		return expressionSqlizer{}, skyerr.NewError(skyerr.RecordQueryInvalid, err.Error())
	}

	field := record.FieldType{}
	for i, keyPathField := range fields {
		isLast := (i == len(components)-1)
		field = keyPathField
		if field.Type == record.TypeReference && !isLast {
			alias = f.createLeftJoin(field.ReferenceType, components[i], "_id")
		}
	}
	return newExpressionSqlizer(alias, field, expr), nil
}

// createLeftJoin create an alias of a table to be joined to the primary table
// and return the alias for the joined table
func (f *predicateSqlizerFactory) createLeftJoin(secondaryTable string, primaryColumn string, secondaryColumn string) string {
	newAlias := joinedTable{secondaryTable, primaryColumn, secondaryColumn}
	for i, alias := range f.joinedTables {
		if alias.equal(newAlias) {
			return f.aliasName(secondaryTable, i)
		}
	}

	f.joinedTables = append(f.joinedTables, newAlias)
	return f.aliasName(secondaryTable, len(f.joinedTables)-1)
}

func (f *predicateSqlizerFactory) aliasName(secondaryTable string, indexInJoinedTables int) string {
	// The _auth table always have the same alias name for
	// getting user info in user discovery
	if secondaryTable == "_auth" {
		return "_auth"
	}
	return fmt.Sprintf("_t%d", indexInJoinedTables)
}

// AddJoinsToSelectBuilder adds join clauses to a SelectBuilder
func (f *predicateSqlizerFactory) AddJoinsToSelectBuilder(q sq.SelectBuilder) sq.SelectBuilder {
	for i, alias := range f.joinedTables {
		aliasName := f.aliasName(alias.secondaryTable, i)
		joinClause := fmt.Sprintf("%s AS %s ON %s = %s",
			f.sqlBuilder.TableName(alias.secondaryTable), pq.QuoteIdentifier(aliasName),
			fullQuoteIdentifier(f.primaryTable, alias.primaryColumn),
			fullQuoteIdentifier(aliasName, alias.secondaryColumn))
		q = q.LeftJoin(joinClause)
	}

	if len(f.joinedTables) > 0 {
		q = q.Distinct()
	}
	return q
}

func (f *predicateSqlizerFactory) addExtraColumn(key string, fieldType record.DataType, expr record.Expression) {
	if f.extraColumns == nil {
		f.extraColumns = map[string]record.FieldType{}
	}
	f.extraColumns[key] = record.FieldType{
		Type:       fieldType,
		Expression: expr,
	}
}

func (f *predicateSqlizerFactory) UpdateTypemap(typemap record.Schema) record.Schema {
	for key, field := range f.extraColumns {
		typemap[key] = field
	}
	return typemap
}

// joinedTable represents a specification for table join
type joinedTable struct {
	secondaryTable  string
	primaryColumn   string
	secondaryColumn string
}

// equal compares whether two specifications of table join are equal
func (a joinedTable) equal(b joinedTable) bool {
	return a.secondaryTable == b.secondaryTable && a.primaryColumn == b.primaryColumn && a.secondaryColumn == b.secondaryColumn
}
