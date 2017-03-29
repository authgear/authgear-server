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

package pq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

const (
	ContextWhere  = "where"
	ContextSelect = "select"
)

// predicateSqlizerFactory is a factory for creating sqlizer for predicate
type predicateSqlizerFactory struct {
	db           *database
	primaryTable string
	joinedTables []joinedTable
	extraColumns map[string]skydb.FieldType
}

func (f *predicateSqlizerFactory) newPredicateSqlizer(p skydb.Predicate) (sq.Sqlizer, error) {
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
			sqlizer, err := f.newPredicateSqlizer(child.(skydb.Predicate))
			if err != nil {
				return nil, err
			}
			and[i] = sqlizer
		}
		return and, nil
	case skydb.Or:
		or := make(sq.Or, len(p.Children))
		for i, child := range p.Children {
			sqlizer, err := f.newPredicateSqlizer(child.(skydb.Predicate))
			if err != nil {
				return nil, err
			}
			or[i] = sqlizer
		}
		return or, nil
	case skydb.Not:
		pred := p.Children[0].(skydb.Predicate)
		sqlizer, err := f.newPredicateSqlizer(pred)
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
				newExpressionSqlizer(alias, lhsExpr),
				newExpressionSqlizer(alias, rhsExpr),
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

func (f *predicateSqlizerFactory) newAccessControlSqlizer(user *skydb.UserInfo, aclLevel skydb.ACLLevel) (sq.Sqlizer, error) {
	return &accessPredicateSqlizer{
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

func newExpressionSqlizer(alias string, expr skydb.Expression) expressionSqlizer {
	return expressionSqlizer{
		alias,
		expr,
		ContextWhere,
		skydb.FieldType{},
	}
}

func (f *predicateSqlizerFactory) newExpressionSqlizer(expr skydb.Expression) (expressionSqlizer, error) {
	if expr.IsKeyPath() {
		return f.newExpressionSqlizerForKeyPath(expr)
	}

	sqlizer := newExpressionSqlizer(
		f.primaryTable,
		expr,
	)

	return sqlizer, nil
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
	for i, component := range components {
		isLast := (i == len(components)-1)

		schema, err := f.db.remoteColumnTypes(recordType)
		if err != nil {
			return expressionSqlizer{}, skyerr.NewErrorf(skyerr.RecordQueryInvalid,
				`record type "%s" does not exist`, recordType)
		}

		field, ok := schema[component]
		if !ok {
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

	return newExpressionSqlizer(alias, expr), nil
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

// addJoinsToSelectBuilder add join clauses to a SelectBuilder
func (f *predicateSqlizerFactory) addJoinsToSelectBuilder(q sq.SelectBuilder) sq.SelectBuilder {
	for i, alias := range f.joinedTables {
		aliasName := f.aliasName(alias.secondaryTable, i)
		joinClause := fmt.Sprintf("%s AS %s ON %s = %s",
			f.db.tableName(alias.secondaryTable), pq.QuoteIdentifier(aliasName),
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

func (f *predicateSqlizerFactory) updateTypemap(typemap skydb.RecordSchema) skydb.RecordSchema {
	for key, field := range f.extraColumns {
		typemap[key] = field
	}
	return typemap
}

func newPredicateSqlizerFactory(db *database, primaryTable string) *predicateSqlizerFactory {
	return &predicateSqlizerFactory{
		db:           db,
		primaryTable: primaryTable,
		joinedTables: []joinedTable{},
	}
}

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
	level skydb.ACLLevel
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

	if lhs.Type == skydb.Literal && rhs.Type == skydb.KeyPath {
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
	} else {
		panic("malformed query")
	}
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

// expressionSqlizer generates an SQL expression from a skydb.Expression. A SQL
// expression are those found in SELECT clause or in the WHERE clause.
//
// In addition to generating literal value such as string (`"hello"`) or integer (`1`),
// the expressionSqlizer can also generate expression for a column or a function.
//
// When expression is a column of a table, the `alias` field is required
// and it is either the name of the table of the column, or a SQL alias of such
// table.
type expressionSqlizer struct {
	alias string
	skydb.Expression
	context   string
	fieldType skydb.FieldType
}

func (expr *expressionSqlizer) ToSql() (sql string, args []interface{}, err error) {
	switch expr.Type {
	case skydb.KeyPath:
		components := expr.KeyPathComponents()
		lastComponent := components[len(components)-1]
		sql = fullQuoteIdentifier(expr.alias, lastComponent)
		args = []interface{}{}
	case skydb.Function:
		sql, args = funcToSQLOperand(expr.alias, expr.Value.(skydb.Func))
	default:
		sql, args = literalToSQLOperand(expr.Value)
	}

	if expr.fieldType.Type == skydb.TypeGeometry && expr.context == ContextSelect {
		sql = selectGeoJSON(sql)
	}

	return
}

func funcToSQLOperand(alias string, fun skydb.Func) (string, []interface{}) {
	switch f := fun.(type) {
	case skydb.DistanceFunc:
		sql := fmt.Sprintf("ST_Distance_Sphere(%s, ST_MakePoint(?, ?))",
			fullQuoteIdentifier(alias, f.Field))
		args := []interface{}{f.Location.Lng(), f.Location.Lat()}
		return sql, args
	case skydb.CountFunc:
		var sql string
		if f.OverallRecords {
			sql = fmt.Sprintf("COUNT(*) OVER()")
		} else {
			sql = fmt.Sprintf("COUNT(*)")
		}
		args := []interface{}{}
		return sql, args
	case skydb.UserDataFunc:
		return fmt.Sprintf("_user.%s", f.DataName), []interface{}{}
	default:
		panic(fmt.Errorf("got unrecgonized skydb.Func = %T", fun))
	}
}

func literalToSQLOperand(literal interface{}) (string, []interface{}) {
	// Array detection is borrowed from squirrel's expr.go
	switch literalValue := literal.(type) {
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
	case skydb.Reference:
		return v.ID.Key
	default:
		return value
	}
}

func sortOrderBySQL(alias string, sort skydb.Sort) (string, error) {
	var expr string

	switch {
	case sort.KeyPath != "":
		expr = fullQuoteIdentifier(alias, sort.KeyPath)
	case sort.Func != nil:
		var err error
		expr, err = funcOrderBySQL(alias, sort.Func)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("invalid Sort: specify either KeyPath or Func")
	}

	order, err := sortOrderOrderBySQL(sort.Order)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(expr + " " + order), nil
}

func selectGeoJSON(sql string) string {
	return fmt.Sprintf("ST_AsGeoJSON(%s)", sql)
}

// due to sq not being able to pass args in OrderBy, we can't re-use funcToSQLOperand
func funcOrderBySQL(alias string, fun skydb.Func) (string, error) {
	switch f := fun.(type) {
	case skydb.DistanceFunc:
		sql := fmt.Sprintf(
			"ST_Distance_Sphere(%s, ST_MakePoint(%f, %f))",
			fullQuoteIdentifier(alias, f.Field),
			f.Location.Lng(),
			f.Location.Lat(),
		)
		return sql, nil
	default:
		return "", fmt.Errorf("got unrecgonized skydb.Func = %T", fun)
	}
}

func sortOrderOrderBySQL(order skydb.SortOrder) (string, error) {
	switch order {
	case skydb.Asc:
		return "ASC", nil
	case skydb.Desc:
		return "DESC", nil
	default:
		return "", fmt.Errorf("unknown sort order = %v", order)
	}
}

func pqDataType(dataType skydb.DataType) string {
	switch dataType {
	default:
		panic(fmt.Sprintf("Unsupported dataType = %s", dataType))
	case skydb.TypeString, skydb.TypeAsset, skydb.TypeReference:
		return TypeString
	case skydb.TypeNumber:
		return TypeNumber
	case skydb.TypeInteger:
		return TypeInteger
	case skydb.TypeDateTime:
		return TypeTimestamp
	case skydb.TypeBoolean:
		return TypeBoolean
	case skydb.TypeJSON:
		return TypeJSON
	case skydb.TypeLocation:
		return TypeLocation
	case skydb.TypeSequence:
		return TypeSerial
	case skydb.TypeGeometry:
		return TypeGeometry
	}
}

func fullQuoteIdentifier(aliasName string, columnName string) string {
	return pq.QuoteIdentifier(aliasName) + "." + pq.QuoteIdentifier(columnName)
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
