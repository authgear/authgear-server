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

	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type PredicateSqlizerFactory interface {
	UpdateTypemap(typemap skydb.RecordSchema) skydb.RecordSchema
	AddJoinsToSelectBuilder(q sq.SelectBuilder) sq.SelectBuilder
	NewPredicateSqlizer(p skydb.Predicate) (sq.Sqlizer, error)
	NewAccessControlSqlizer(user *skydb.UserInfo, aclLevel skydb.RecordACLLevel) (sq.Sqlizer, error)
}

// predicateSqlizerFactory is a factory for creating sqlizer for predicate
type predicateSqlizerFactory struct {
	db           skydb.Database
	primaryTable string
	joinedTables []joinedTable
	extraColumns map[string]skydb.FieldType
}

func NewPredicateSqlizerFactory(db skydb.Database, primaryTable string) PredicateSqlizerFactory {
	return &predicateSqlizerFactory{
		db:           db,
		primaryTable: primaryTable,
		joinedTables: []joinedTable{},
	}
}

func (f *predicateSqlizerFactory) NewPredicateSqlizer(p skydb.Predicate) (sq.Sqlizer, error) {
	if p.IsEmpty() {
		panic("no sqlizer can be created from an empty predicate")
	}

	if p.Operator == skydb.Functional {
		return f.newFunctionalPredicateSqlizer(p)
	}
	if p.Operator.IsCompound() {
		return f.newCompoundPredicateSqlizer(p)
	}
	return f.newComparisonPredicateSqlizer(p)
}

func (f *predicateSqlizerFactory) newCompoundPredicateSqlizer(p skydb.Predicate) (sq.Sqlizer, error) {
	switch p.Operator {
	default:
		err := fmt.Errorf("compound operator `%v` is not supported", p.Operator)
		return nil, err
	case skydb.And:
		and := make(sq.And, len(p.Children))
		for i, child := range p.Children {
			sqlizer, err := f.NewPredicateSqlizer(child.(skydb.Predicate))
			if err != nil {
				return nil, err
			}
			and[i] = sqlizer
		}
		return and, nil
	case skydb.Or:
		or := make(sq.Or, len(p.Children))
		for i, child := range p.Children {
			sqlizer, err := f.NewPredicateSqlizer(child.(skydb.Predicate))
			if err != nil {
				return nil, err
			}
			or[i] = sqlizer
		}
		return or, nil
	case skydb.Not:
		pred := p.Children[0].(skydb.Predicate)
		sqlizer, err := f.NewPredicateSqlizer(pred)
		if err != nil {
			return nil, err
		}
		return NotSqlizer{sqlizer}, nil
	}
}

func (f *predicateSqlizerFactory) newFunctionalPredicateSqlizer(predicate skydb.Predicate) (sq.Sqlizer, error) {
	expr := predicate.Children[0].(skydb.Expression)
	if expr.Type != skydb.Function {
		panic("unexpected expression in functional predicate")
	}
	switch fn := expr.Value.(type) {
	case skydb.UserRelationFunc:
		return f.newUserRelationFunctionalPredicateSqlizer(fn)
	case skydb.UserDiscoverFunc:
		return f.newUserDiscoverFunctionalPredicateSqlizer(fn)
	default:
		panic("the specified function cannot be used as a functional predicate")
	}
}

func (f *predicateSqlizerFactory) newUserRelationFunctionalPredicateSqlizer(fn skydb.UserRelationFunc) (sq.Sqlizer, error) {
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

func (f *predicateSqlizerFactory) newUserDiscoverFunctionalPredicateSqlizer(fn skydb.UserDiscoverFunc) (sq.Sqlizer, error) {
	if f.db.UserRecordType() != f.primaryTable {
		return nil, skyerr.NewErrorf(skyerr.RecordQueryInvalid,
			"user discover predicate can only be used on user record")
	}

	discoveryArgNames := []string{"username", "email"}
	var sqlizers sq.Or
	var alias string

	// Create sqlizers for each discovery argument (username and email).
	for _, argName := range discoveryArgNames {
		if !fn.HaveArgsByName(argName) {
			continue
		}

		lhsExpr := skydb.Expression{
			Type:  skydb.KeyPath,
			Value: argName,
		}
		rhsExpr := skydb.Expression{
			Type:  skydb.Literal,
			Value: fn.ArgsByName(argName),
		}

		if alias == "" {
			alias = f.createLeftJoin("_user", "_id", "id")
		}
		sqlizer := &containsComparisonPredicateSqlizer{
			[]expressionSqlizer{
				newExpressionSqlizer(alias, skydb.FieldType{Type: skydb.TypeString}, lhsExpr),
				newExpressionSqlizer(alias, skydb.FieldType{Type: skydb.TypeString}, rhsExpr),
			},
		}
		sqlizers = append(sqlizers, sqlizer)
	}

	// If there are no sqlizers, we return early with a
	// sqlizers that always evaluates to false.
	if len(sqlizers) == 0 {
		return FalseSqlizer{}, nil
	}

	// Add transient attributes so that returned record also contain
	// username and email.
	for _, argName := range discoveryArgNames {
		transientColumn := fmt.Sprintf("_transient__%s", argName)

		expr := skydb.Expression{
			Type:  skydb.Function,
			Value: skydb.UserDataFunc{argName},
		}
		f.addExtraColumn(transientColumn, skydb.TypeString, expr)
	}

	return sqlizers, nil
}

func (f *predicateSqlizerFactory) NewAccessControlSqlizer(user *skydb.UserInfo, aclLevel skydb.RecordACLLevel) (sq.Sqlizer, error) {
	return &accessPredicateSqlizer{
		f.primaryTable,
		user,
		aclLevel,
	}, nil
}

func (f *predicateSqlizerFactory) newComparisonPredicateSqlizer(p skydb.Predicate) (sq.Sqlizer, error) {
	if sqlizer, ok := f.tryOptimizeDistancePredicate(p); ok {
		return sqlizer, nil
	}

	sqlizers := []expressionSqlizer{}
	for _, child := range p.Children {
		sqlizer, err := f.newExpressionSqlizer(child.(skydb.Expression))
		if err != nil {
			return nil, err
		}
		sqlizers = append(sqlizers, sqlizer)
	}

	if p.Operator == skydb.In {
		return &containsComparisonPredicateSqlizer{sqlizers}, nil
	}
	return &comparisonPredicateSqlizer{sqlizers, p.Operator}, nil
}

// tryOptimizeDistancePredicate returns a sqlizer that is more efficient
// at querying whether two points are within certain distance.
//
// If the predicate cannot be optimize or an error occurred generating
// an optimized sqlizer, the second value returned is false.
func (f *predicateSqlizerFactory) tryOptimizeDistancePredicate(p skydb.Predicate) (sq.Sqlizer, bool) {
	var tryFunc skydb.Expression
	var tryValue skydb.Expression
	if p.Operator == skydb.LessThan {
		tryFunc = p.Children[0].(skydb.Expression)
		tryValue = p.Children[1].(skydb.Expression)
	} else if p.Operator == skydb.GreaterThan {
		tryFunc = p.Children[1].(skydb.Expression)
		tryValue = p.Children[0].(skydb.Expression)
	} else {
		return nil, false
	}

	distanceFunc, ok := tryFunc.Value.(skydb.DistanceFunc)
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

func (f *predicateSqlizerFactory) newExpressionSqlizer(expr skydb.Expression) (expressionSqlizer, error) {
	if expr.IsKeyPath() {
		return f.newExpressionSqlizerForKeyPath(expr)
	}

	if expr.Type == skydb.Literal {
		var fieldType skydb.FieldType
		if expr.Value != nil {
			var err error
			fieldType, err = skydb.DeriveFieldType(expr.Value)
			if err != nil {
				return expressionSqlizer{}, err
			}
		}

		sqlizer := newExpressionSqlizer(f.primaryTable, fieldType, expr)
		return sqlizer, nil
	}

	if expr.Type == skydb.Function {
		funcInterface, ok := expr.Value.(skydb.Func)
		if !ok {
			panic(`expression value is not a function`)
		}
		return newExpressionSqlizer(f.primaryTable, skydb.FieldType{Type: funcInterface.DataType()}, expr), nil
	}

	return expressionSqlizer{}, skyerr.NewError(skyerr.RecordQueryInvalid,
		`unexpected expression type`)
}

func (f *predicateSqlizerFactory) newExpressionSqlizerForKeyPath(expr skydb.Expression) (expressionSqlizer, error) {
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
	recordType := f.primaryTable
	field := skydb.FieldType{}
	for i, component := range components {
		isLast := (i == len(components)-1)

		schema, err := f.db.RemoteColumnTypes(recordType)
		if err != nil {
			return expressionSqlizer{}, skyerr.NewErrorf(skyerr.RecordQueryInvalid,
				`record type "%s" does not exist`, recordType)
		}

		if f, ok := schema[component]; ok {
			field = f
		} else {
			return expressionSqlizer{}, skyerr.NewErrorf(skyerr.RecordQueryInvalid,
				`keypath "%s" does not exist`, keyPath)
		}

		if field.Type != skydb.TypeReference && !isLast {
			return expressionSqlizer{}, skyerr.NewErrorf(skyerr.RecordQueryInvalid,
				`field "%s" in keypath "%s" is not a reference`, component, keyPath)
		}

		if field.Type == skydb.TypeReference && !isLast {
			// follow the keypath and join the table
			alias = f.createLeftJoin(field.ReferenceType, component, "_id")
			recordType = field.ReferenceType
			continue
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
	// The _user table always have the same alias name for
	// getting user info in user discovery
	if secondaryTable == "_user" {
		return "_user"
	}
	return fmt.Sprintf("_t%d", indexInJoinedTables)
}

// AddJoinsToSelectBuilder adds join clauses to a SelectBuilder
func (f *predicateSqlizerFactory) AddJoinsToSelectBuilder(q sq.SelectBuilder) sq.SelectBuilder {
	for i, alias := range f.joinedTables {
		aliasName := f.aliasName(alias.secondaryTable, i)
		joinClause := fmt.Sprintf("%s AS %s ON %s = %s",
			f.db.TableName(alias.secondaryTable), pq.QuoteIdentifier(aliasName),
			fullQuoteIdentifier(f.primaryTable, alias.primaryColumn),
			fullQuoteIdentifier(aliasName, alias.secondaryColumn))
		q = q.LeftJoin(joinClause)
	}

	if len(f.joinedTables) > 0 {
		q = q.Distinct()
	}
	return q
}

func (f *predicateSqlizerFactory) addExtraColumn(key string, fieldType skydb.DataType, expr skydb.Expression) {
	if f.extraColumns == nil {
		f.extraColumns = map[string]skydb.FieldType{}
	}
	f.extraColumns[key] = skydb.FieldType{
		Type:       fieldType,
		Expression: expr,
	}
}

func (f *predicateSqlizerFactory) UpdateTypemap(typemap skydb.RecordSchema) skydb.RecordSchema {
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
