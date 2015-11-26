package pq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/oursky/skygear/skydb"
	"github.com/paulmach/go.geo"
)

// This file implements Record related operations of the
// skydb/pq implementation.

// Different data types that can be saved in and loaded from postgreSQL
// NOTE(limouren): varchar is missing because text can replace them,
// see the docs here: http://www.postgresql.org/docs/9.4/static/datatype-character.html
const (
	TypeString    = "text"
	TypeNumber    = "double precision"
	TypeBoolean   = "boolean"
	TypeJSON      = "jsonb"
	TypeTimestamp = "timestamp without time zone"
	TypeLocation  = "geometry(Point)"
	TypeInteger   = "integer"
	TypeSerial    = "serial UNIQUE"
)

type nullJSON struct {
	JSON  interface{}
	Valid bool
}

func (nj *nullJSON) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if value == nil || !ok {
		nj.JSON = nil
		nj.Valid = false
		return nil
	}

	err := json.Unmarshal(data, &nj.JSON)
	nj.Valid = err == nil

	return err
}

type assetValue skydb.Asset

func (asset assetValue) Value() (driver.Value, error) {
	return asset.Name, nil
}

type nullAsset struct {
	Asset *skydb.Asset
	Valid bool
}

func (na *nullAsset) Scan(value interface{}) error {
	if value == nil {
		na.Asset = &skydb.Asset{}
		na.Valid = false
		return nil
	}

	assetName, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Asset: got type(value) = %T, expect []byte", value)
	}

	na.Asset = &skydb.Asset{
		Name: string(assetName),
	}
	na.Valid = true

	return nil
}

type nullLocation struct {
	Location skydb.Location
	Valid    bool
}

func (nl *nullLocation) Scan(value interface{}) error {
	if value == nil {
		nl.Location = skydb.Location{}
		nl.Valid = false
		return nil
	}

	src, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Location: got type(value) = %T, expect []byte", value)
	}

	// TODO(limouren): instead of decoding a str-encoded hex, we should utilize
	// ST_AsBinary to perform the SELECT
	decoded := make([]byte, hex.DecodedLen(len(src)))
	_, err := hex.Decode(decoded, src)
	if err != nil {
		return fmt.Errorf("failed to scan Location: malformed wkb")
	}

	err = (*geo.Point)(&nl.Location).Scan(decoded)
	nl.Valid = err == nil
	return err
}

type referenceValue skydb.Reference

func (ref referenceValue) Value() (driver.Value, error) {
	return ref.ID.Key, nil
}

type jsonSliceValue []interface{}

func (s jsonSliceValue) Value() (driver.Value, error) {
	return json.Marshal([]interface{}(s))
}

type jsonMapValue map[string]interface{}

func (m jsonMapValue) Value() (driver.Value, error) {
	return json.Marshal(map[string]interface{}(m))
}

type aclValue skydb.RecordACL

func (acl aclValue) Value() (driver.Value, error) {
	if acl == nil {
		return nil, nil
	}
	return json.Marshal(acl)
}

type locationValue skydb.Location

func (loc *locationValue) Value() (driver.Value, error) {
	return (*geo.Point)(loc).ToWKT(), nil
}

func (db *database) Get(id skydb.RecordID, record *skydb.Record) error {
	typemap, err := db.remoteColumnTypes(id.Type)
	if err != nil {
		return err
	}

	if len(typemap) == 0 { // record type has not been created
		return skydb.ErrRecordNotFound
	}

	sqlStmt, args, err := db.selectQuery(id.Type, typemap).Where("_id = ?", id.Key).ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sqlStmt,
		"args": args,
	}).Debugln("Getting record")

	row := db.Db.QueryRowx(sqlStmt, args...)
	if err := newRecordScanner(id.Type, typemap, row).Scan(record); err == sql.ErrNoRows {
		return skydb.ErrRecordNotFound
	} else if err != nil {
		return err
	}
	return nil
}

// Save attempts to do a upsert
func (db *database) Save(record *skydb.Record) error {
	if record.ID.Key == "" {
		return errors.New("db.save: got empty record id")
	}
	if record.ID.Type == "" {
		return fmt.Errorf("db.save %s: got empty record type", record.ID.Key)
	}
	if record.OwnerID == "" {
		return fmt.Errorf("db.save %s: got empty OwnerID", record.ID.Key)
	}

	pkData := map[string]interface{}{
		"_id":          record.ID.Key,
		"_database_id": db.userID,
	}
	upsert := upsertQuery(db.tableName(record.ID.Type), pkData, convert(record)).
		IgnoreKeyOnUpdate("_owner_id").
		IgnoreKeyOnUpdate("_created_at").
		IgnoreKeyOnUpdate("_created_by")

	typemap, err := db.remoteColumnTypes(record.ID.Type)
	if err != nil {
		return err
	}

	if err := db.preSave(typemap, record); err != nil {
		return err
	}

	row := queryRowWith(db.Db, upsert)
	if err = newRecordScanner(record.ID.Type, typemap, row).Scan(record); err != nil {
		sql, args, _ := upsert.ToSql()
		log.WithFields(log.Fields{
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Debugln("Failed to save record")

		return err
	}

	record.DatabaseID = db.userID
	return nil
}

func (db *database) preSave(schema skydb.RecordSchema, record *skydb.Record) error {
	const SetSequenceMaxValue = `SELECT setval($1, GREATEST(max(%v), $2)) FROM %v;`

	for key, value := range record.Data {
		// we are setting a sequence field
		if schema[key].Type == skydb.TypeInteger {
			selectSQL := fmt.Sprintf(SetSequenceMaxValue, pq.QuoteIdentifier(key), db.tableName(record.ID.Type))
			seqName := db.tableName(fmt.Sprintf(`%v_%v_seq`, record.ID.Type, key))
			if _, err := db.Db.Exec(selectSQL, seqName, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func convert(r *skydb.Record) map[string]interface{} {
	m := map[string]interface{}{}
	for key, rawValue := range r.Data {
		switch value := rawValue.(type) {
		case []interface{}:
			m[key] = jsonSliceValue(value)
		case map[string]interface{}:
			m[key] = jsonMapValue(value)
		case *skydb.Asset:
			m[key] = assetValue(*value)
		case skydb.Reference:
			m[key] = referenceValue(value)
		case *skydb.Location:
			m[key] = (*locationValue)(value)
		default:
			m[key] = rawValue
		}
	}
	m["_owner_id"] = r.OwnerID
	m["_access"] = aclValue(r.ACL)
	m["_created_at"] = r.CreatedAt
	m["_created_by"] = r.CreatorID
	m["_updated_at"] = r.UpdatedAt
	m["_updated_by"] = r.UpdaterID
	return m
}

func (db *database) Delete(id skydb.RecordID) error {
	builder := psql.Delete(db.tableName(id.Type)).
		Where("_id = ? AND _database_id = ?", id.Key, db.userID)

	result, err := execWith(db.Db, builder)
	if isUndefinedTable(err) {
		return skydb.ErrRecordNotFound
	} else if err != nil {
		sql, args, _ := builder.ToSql()
		log.WithFields(log.Fields{
			"id":   id,
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Errorln("Failed to execute delete record statement")
		return fmt.Errorf("delete %s: failed to delete record", id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		sql, args, _ := builder.ToSql()
		log.WithFields(log.Fields{
			"id":   id,
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Errorln("Failed to fetch rowsAffected")
		return fmt.Errorf("delete %s: failed to retrieve deletion status", id)
	}

	if rowsAffected == 0 {
		return skydb.ErrRecordNotFound
	} else if rowsAffected > 1 {
		sql, args, _ := builder.ToSql()
		log.WithFields(log.Fields{
			"id":           id,
			"sql":          sql,
			"args":         args,
			"rowsAffected": rowsAffected,
			"err":          err,
		}).Errorln("Unexpected rows deleted")
		return fmt.Errorf("delete %s: got %v rows deleted, want 1", id, rowsAffected)
	}

	return err
}

type comparisonPredicateSqlizer struct {
	alias string
	skydb.Predicate
}

type containsComparisonPredicateSqlizer struct {
	alias string
	skydb.Predicate
}

// predicateSqlizerFactory is a factory for creating sqlizer for predicate
type predicateSqlizerFactory struct {
	db           *database
	primaryTable string
	joinedTables []joinedTable
}

func (f *predicateSqlizerFactory) newPredicateSqlizer(predicate skydb.Predicate) (sq.Sqlizer, error) {
	if predicate.Operator == skydb.Functional {
		return f.newFunctionalPredicateSqlizer(predicate)
	}
	if predicate.Operator.IsCompound() {
		return f.newCompoundPredicateSqlizer(predicate)
	}
	if predicate.Operator == skydb.In {
		return &containsComparisonPredicateSqlizer{f.primaryTable, predicate}, nil
	}
	return &comparisonPredicateSqlizer{f.primaryTable, predicate}, nil
}

func (f *predicateSqlizerFactory) newCompoundPredicateSqlizer(p skydb.Predicate) (sq.Sqlizer, error) {
	switch p.Operator {
	default:
		err := fmt.Errorf("Compound operator `%v` is not supported.", p.Operator)
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

type expressionSqlizer struct {
	alias string
	skydb.Expression
}

func (expr *expressionSqlizer) ToSql() (sql string, args []interface{}, err error) {
	switch expr.Type {
	case skydb.KeyPath:
		sql = fullQuoteIdentifier(expr.alias, expr.Value.(string))
		args = []interface{}{}
	case skydb.Function:
		sql, args = funcToSQLOperand(expr.alias, expr.Value.(skydb.Func))
	default:
		sql, args = literalToSQLOperand(expr.Value)
	}
	return
}

func funcToSQLOperand(alias string, fun skydb.Func) (string, []interface{}) {
	switch f := fun.(type) {
	case *skydb.DistanceFunc:
		sql := fmt.Sprintf("ST_Distance_Sphere(%s, ST_MakePoint(?, ?))",
			fullQuoteIdentifier(alias, f.Field))
		args := []interface{}{f.Location.Lng(), f.Location.Lat()}
		return sql, args
	case *skydb.CountFunc:
		var sql string
		if f.OverallRecords {
			sql = fmt.Sprintf("COUNT(*) OVER()")
		} else {
			sql = fmt.Sprintf("COUNT(*)")
		}
		args := []interface{}{}
		return sql, args
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

func (f *predicateSqlizerFactory) newFunctionalPredicateSqlizer(predicate skydb.Predicate) (sq.Sqlizer, error) {
	expr := predicate.Children[0].(skydb.Expression)
	if expr.Type != skydb.Function {
		panic("unexpected expression in functional predicate")
	}
	switch fn := expr.Value.(type) {
	case *skydb.UserRelationFunc:
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
	default:
		panic("the specified function cannot be used as a functional predicate")
	}
}

// createLeftJoin create an alias of a table to be joined to the primary table
// and return the alias for the joined table
func (f *predicateSqlizerFactory) createLeftJoin(secondaryTable string, primaryColumn string, secondaryColumn string) string {
	newAlias := joinedTable{secondaryTable, primaryColumn, secondaryColumn}
	for i, alias := range f.joinedTables {
		if alias.equal(newAlias) {
			return fmt.Sprintf("_t%d", i)
		}
	}

	f.joinedTables = append(f.joinedTables, newAlias)
	return fmt.Sprintf("_t%d", len(f.joinedTables)-1)
}

// addJoinsToSelectBuilder add join clauses to a SelectBuilder
func (f *predicateSqlizerFactory) addJoinsToSelectBuilder(q sq.SelectBuilder) sq.SelectBuilder {
	for i, alias := range f.joinedTables {
		aliasName := fmt.Sprintf("_t%d", i)
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

func newPredicateSqlizerFactory(db *database, primaryTable string) *predicateSqlizerFactory {
	return &predicateSqlizerFactory{
		db:           db,
		primaryTable: primaryTable,
		joinedTables: []joinedTable{},
	}
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

func (p *containsComparisonPredicateSqlizer) ToSql() (sql string, args []interface{}, err error) {
	var buffer bytes.Buffer
	lhs := expressionSqlizer{p.alias, p.Children[0].(skydb.Expression)}
	rhs := expressionSqlizer{p.alias, p.Children[1].(skydb.Expression)}

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

func (p *comparisonPredicateSqlizer) ToSql() (sql string, args []interface{}, err error) {
	args = []interface{}{}
	if p.Operator.IsBinary() {
		var buffer bytes.Buffer
		lhs := expressionSqlizer{p.alias, p.Children[0].(skydb.Expression)}
		rhs := expressionSqlizer{p.alias, p.Children[1].(skydb.Expression)}

		sqlOperand, opArgs, err := lhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		switch p.Operator {
		default:
			err = fmt.Errorf("Comparison operator `%v` is not supported.", p.Operator)
			return sql, args, err
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

		sqlOperand, opArgs, err = rhs.ToSql()
		if err != nil {
			return "", nil, err
		}
		buffer.WriteString(sqlOperand)
		args = append(args, opArgs...)

		sql = buffer.String()
	} else {
		err = fmt.Errorf("Comparison operator `%v` is not supported.", p.Operator)
	}

	return
}

func (db *database) Query(query *skydb.Query) (*skydb.Rows, error) {
	if query.Type == "" {
		return nil, errors.New("got empty query type")
	}

	typemap, err := db.remoteColumnTypes(query.Type)
	if err != nil {
		return nil, err
	}

	if len(typemap) == 0 { // record type has not been created
		return skydb.EmptyRows, nil
	}

	if query.DesiredKeys != nil {
		newtypemap, err := whitelistedRecordSchema(typemap, query.DesiredKeys)
		if err != nil {
			return nil, err
		}
		typemap = newtypemap
	}

	for key, value := range query.ComputedKeys {
		if value.Type == skydb.KeyPath {
			// recorddb does not support querying with computed keys
			continue
		}

		typemap["_transient_"+key] = skydb.FieldType{
			Type:       skydb.TypeNumber,
			Expression: &value,
		}
	}

	if query.GetCount {
		typemap["_record_count"] = skydb.FieldType{
			Type: skydb.TypeNumber,
			Expression: &skydb.Expression{
				Type: skydb.Function,
				Value: &skydb.CountFunc{
					OverallRecords: true,
				},
			},
		}
	}

	q := db.selectQuery(query.Type, typemap)

	if p := query.Predicate; p != nil {
		factory := newPredicateSqlizerFactory(db, query.Type)
		sqlizer, err := factory.newPredicateSqlizer(*p)
		if err != nil {
			return nil, err
		}
		q = q.Where(sqlizer)
		q = factory.addJoinsToSelectBuilder(q)
	}

	for _, sort := range query.Sorts {
		orderBy, err := sortOrderBySQL(query.Type, sort)
		if err != nil {
			return nil, err
		}
		q = q.OrderBy(orderBy)
	}

	if query.ReadableBy != "" {
		// FIXME: Serialize the json instead of building manually
		q = q.Where(
			`(_access @> '[{"user_id":"`+query.ReadableBy+`"}]' OR `+
				`_access IS NULL OR `+
				`_owner_id = ?)`, query.ReadableBy)
	}

	if query.Limit != nil {
		q = q.Limit(*query.Limit)
	}

	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}

	sql, args, err := q.ToSql()
	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
		"err":  err,
	}).Infoln("query records")

	rows, err := queryWith(db.Db, q)
	return newRows(query.Type, typemap, rows, err)
}

func (db *database) QueryCount(query *skydb.Query) (uint64, error) {
	if query.Type == "" {
		return 0, errors.New("got empty query type")
	}

	typemap, err := db.remoteColumnTypes(query.Type)
	if err != nil || len(typemap) == 0 { // error or record type has not been created
		return 0, err
	}

	typemap = skydb.RecordSchema{
		"_record_count": skydb.FieldType{
			Type: skydb.TypeNumber,
			Expression: &skydb.Expression{
				Type: skydb.Function,
				Value: &skydb.CountFunc{
					OverallRecords: false,
				},
			},
		},
	}

	q := db.selectQuery(query.Type, typemap)

	if p := query.Predicate; p != nil {
		factory := newPredicateSqlizerFactory(db, query.Type)
		sqlizer, err := factory.newPredicateSqlizer(*p)
		if err != nil {
			return 0, err
		}
		q = q.Where(sqlizer)
		q = factory.addJoinsToSelectBuilder(q)
	}

	if query.ReadableBy != "" {
		// FIXME: Serialize the json instead of building manually
		q = q.Where(
			`(_access @> '[{"user_id":"`+query.ReadableBy+`"}]' OR `+
				`_access IS NULL OR `+
				`_owner_id = ?)`, query.ReadableBy)
	}

	rows, err := queryWith(db.Db, q)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		panic("Unexpected zero rows returned for aggregate count function.")
	}

	var recordCount uint64
	err = rows.Scan(&recordCount)
	if err != nil {
		return 0, err
	}

	return recordCount, nil
}

func whitelistedRecordSchema(schema skydb.RecordSchema, whitelistKeys []string) (skydb.RecordSchema, error) {
	wlSchema := skydb.RecordSchema{}

	for _, key := range whitelistKeys {
		columnType, ok := schema[key]
		if !ok {
			return nil, fmt.Errorf(`unexpected key "%s"`, key)
		}
		wlSchema[key] = columnType
	}
	for key, value := range schema {
		if strings.HasPrefix(key, "_") {
			wlSchema[key] = value
		}
	}

	return wlSchema, nil
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

// due to sq not being able to pass args in OrderBy, we can't re-use funcToSQLOperand
func funcOrderBySQL(alias string, fun skydb.Func) (string, error) {
	switch f := fun.(type) {
	case *skydb.DistanceFunc:
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

// columnsScanner wraps over sqlx.Rows and sqlx.Row to provide
// a consistent interface for column scanning.
type columnsScanner interface {
	Columns() ([]string, error)
	Scan(dest ...interface{}) error
}

type recordScanner struct {
	recordType  string
	typemap     skydb.RecordSchema
	cs          columnsScanner
	columns     []string
	err         error
	recordCount *uint64
}

func newRecordScanner(recordType string, typemap skydb.RecordSchema, cs columnsScanner) *recordScanner {
	columns, err := cs.Columns()
	return &recordScanner{recordType, typemap, cs, columns, err, nil}
}

func (rs *recordScanner) Scan(record *skydb.Record) error {
	if rs.err != nil {
		return rs.err
	}

	values := make([]interface{}, 0, len(rs.columns))
	for _, column := range rs.columns {
		schema, ok := rs.typemap[column]
		if !ok {
			return fmt.Errorf("received unknown column = %s", column)
		}
		switch schema.Type {
		case skydb.TypeNumber:
			var number sql.NullFloat64
			values = append(values, &number)
		case skydb.TypeString, skydb.TypeReference, skydb.TypeACL:
			var str sql.NullString
			values = append(values, &str)
		case skydb.TypeDateTime:
			var ts pq.NullTime
			values = append(values, &ts)
		case skydb.TypeBoolean:
			var boolean sql.NullBool
			values = append(values, &boolean)
		case skydb.TypeAsset:
			var asset nullAsset
			values = append(values, &asset)
		case skydb.TypeJSON:
			var j nullJSON
			values = append(values, &j)
		case skydb.TypeLocation:
			var l nullLocation
			values = append(values, &l)
		case skydb.TypeInteger:
			var i sql.NullInt64
			values = append(values, &i)
		default:
			return fmt.Errorf("received unknown data type = %v for column = %s", schema.Type, column)
		}
	}

	if err := rs.cs.Scan(values...); err != nil {
		rs.err = err
		return err
	}

	record.ID.Type = rs.recordType
	record.Data = map[string]interface{}{}

	for i, column := range rs.columns {
		value := values[i]

		if column == "_record_count" {
			svalue, ok := value.(*sql.NullFloat64)
			if !ok || !svalue.Valid {
				panic("Unexpected missing column or column is null for _record_count.")
			}

			rs.recordCount = new(uint64)
			*rs.recordCount = uint64(svalue.Float64)
			continue
		}

		switch svalue := value.(type) {
		default:
			return fmt.Errorf("received unexpected scanned type = %T for column = %s", value, column)
		case *sql.NullFloat64:
			if svalue.Valid {
				record.Set(column, svalue.Float64)
			}
		case *sql.NullString:
			if svalue.Valid {
				schema := rs.typemap[column]
				if schema.Type == skydb.TypeReference {
					record.Set(column, skydb.NewReference(schema.ReferenceType, svalue.String))
				} else if schema.Type == skydb.TypeACL {
					acl := skydb.RecordACL{}
					json.Unmarshal([]byte(svalue.String), &acl)
					record.Set(column, acl)
				} else {
					record.Set(column, svalue.String)
				}
			}
		case *pq.NullTime:
			if svalue.Valid {
				// it is to support direct deep-equal of value between
				// a empty record and a record materialized from the database
				if svalue.Time.IsZero() {
					record.Set(column, time.Time{})
				} else {
					record.Set(column, svalue.Time.In(time.UTC))
				}
			}
		case *sql.NullBool:
			if svalue.Valid {
				record.Set(column, svalue.Bool)
			}
		case *nullAsset:
			if svalue.Valid {
				record.Set(column, svalue.Asset)
			}
		case *nullJSON:
			if svalue.Valid {
				record.Set(column, svalue.JSON)
			}
		case *nullLocation:
			if svalue.Valid {
				record.Set(column, &svalue.Location)
			}
		case *sql.NullInt64:
			if svalue.Valid {
				record.Set(column, svalue.Int64)
			}
		}

	}

	return nil
}

type rowsIter struct {
	rows *sqlx.Rows
	rs   *recordScanner
}

func (rowsi rowsIter) Close() error {
	return rowsi.rows.Close()
}

func (rowsi rowsIter) Next(record *skydb.Record) error {
	if rowsi.rows.Next() {
		return rowsi.rs.Scan(record)
	} else if rowsi.rows.Err() != nil {
		return rowsi.rows.Err()
	} else {
		return io.EOF
	}
}

func (rowsi rowsIter) OverallRecordCount() *uint64 {
	return rowsi.rs.recordCount
}

func newRows(recordType string, typemap skydb.RecordSchema, rows *sqlx.Rows, err error) (*skydb.Rows, error) {
	if err != nil {
		return nil, err
	}
	rs := newRecordScanner(recordType, typemap, rows)
	return skydb.NewRows(rowsIter{rows, rs}), nil
}

func (db *database) selectQuery(recordType string, typemap skydb.RecordSchema) sq.SelectBuilder {
	q := psql.Select()
	for column, fieldType := range typemap {
		expr := fieldType.Expression
		if expr == nil {
			expr = &skydb.Expression{
				Type:  skydb.KeyPath,
				Value: column,
			}
		}

		e := expressionSqlizer{recordType, *expr}
		sqlOperand, opArgs, _ := e.ToSql()
		q = q.Column(sqlOperand+" as "+pq.QuoteIdentifier(column), opArgs...)
	}

	q = q.From(db.tableName(recordType)).
		Where("_database_id = ?", db.userID)

	return q
}

// STEP 1 & 2 are obtained by reverse engineering psql \d with -E option
//
// STEP 3: example of getting foreign keys
// SELECT
//     tc.table_name, kcu.column_name,
//     ccu.table_name AS foreign_table_name,
//     ccu.column_name AS foreign_column_name
// FROM
//     information_schema.table_constraints AS tc
//     JOIN information_schema.key_column_usage
//         AS kcu ON tc.constraint_name = kcu.constraint_name
//     JOIN information_schema.constraint_column_usage
//         AS ccu ON ccu.constraint_name = tc.constraint_name
// WHERE constraint_type = 'FOREIGN KEY'
// AND tc.table_schema = 'app__'
// AND tc.table_name = 'note';
func (db *database) remoteColumnTypes(recordType string) (skydb.RecordSchema, error) {
	typemap := skydb.RecordSchema{}
	// STEP 0: Return the cached ColumnType
	if schema, ok := db.c.RecordSchema[recordType]; ok {
		return schema, nil
	}
	defer func() {
		db.c.RecordSchema[recordType] = typemap
		log.Debugf("Cache remoteColumnTypes %s", recordType)
	}()
	log.Debugf("Querying remoteColumnTypes %s", recordType)
	// STEP 1: Get the oid of the current table
	var oid int
	err := db.Db.QueryRowx(`
SELECT c.oid
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE c.relname = $1
  AND n.nspname = $2`,
		recordType, db.schemaName()).Scan(&oid)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.WithFields(log.Fields{
			"schemaName": db.schemaName(),
			"recordType": recordType,
			"err":        err,
		}).Errorln("Failed to query oid of table")
		return nil, err
	}

	// STEP 2: Get column name and data type
	rows, err := db.Db.Queryx(`
SELECT a.attname,
  pg_catalog.format_type(a.atttypid, a.atttypmod)
FROM pg_catalog.pg_attribute a
WHERE a.attrelid = $1 AND a.attnum > 0 AND NOT a.attisdropped`,
		oid)

	if err != nil {
		log.WithFields(log.Fields{
			"schemaName": db.schemaName(),
			"recordType": recordType,
			"oid":        oid,
			"err":        err,
		}).Errorln("Failed to query column and data type")
		return nil, err
	}

	var columnName, pqType string
	for rows.Next() {
		if err := rows.Scan(&columnName, &pqType); err != nil {
			return nil, err
		}

		schema := skydb.FieldType{}
		switch pqType {
		case TypeString:
			schema.Type = skydb.TypeString
		case TypeNumber:
			schema.Type = skydb.TypeNumber
		case TypeTimestamp:
			schema.Type = skydb.TypeDateTime
		case TypeBoolean:
			schema.Type = skydb.TypeBoolean
		case TypeJSON:
			if columnName == "_access" {
				schema.Type = skydb.TypeACL
			} else {
				schema.Type = skydb.TypeJSON
			}
		case TypeLocation:
			schema.Type = skydb.TypeLocation
		case TypeInteger:
			schema.Type = skydb.TypeInteger
		default:
			return nil, fmt.Errorf("received unknown data type = %s for column = %s", pqType, columnName)
		}

		typemap[columnName] = schema
	}

	// STEP 3: FOREIGN KEY, assumeing we can only reference _id i.e. "ccu.column_name" = _id
	builder := psql.Select("kcu.column_name", "ccu.table_name").
		From("information_schema.table_constraints AS tc").
		Join("information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name").
		Join("information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name").
		Where("constraint_type = 'FOREIGN KEY' AND tc.table_schema = ? AND tc.table_name = ?", db.schemaName(), recordType)

	refs, err := queryWith(db.Db, builder)
	if err != nil {
		log.WithFields(log.Fields{
			"schemaName": db.schemaName(),
			"recordType": recordType,
			"err":        err,
		}).Errorln("Failed to query foreign key information schema")

		return nil, err
	}

	for refs.Next() {
		s := skydb.FieldType{}
		var primaryColumn, referencedTable string
		if err := refs.Scan(&primaryColumn, &referencedTable); err != nil {
			log.Debugf("err %v", err)
			return nil, err
		}
		switch referencedTable {
		case "_asset":
			s.Type = skydb.TypeAsset
		default:
			s.Type = skydb.TypeReference
			s.ReferenceType = referencedTable
		}
		typemap[primaryColumn] = s
	}
	return typemap, nil
}

func (db *database) Extend(recordType string, recordSchema skydb.RecordSchema) error {
	remoteRecordSchema, err := db.remoteColumnTypes(recordType)
	if err != nil {
		return err
	}

	if len(remoteRecordSchema) == 0 {
		if err := db.createTable(recordType); err != nil {
			return fmt.Errorf("failed to create table: %s", err)
		}
	}
	updatingSchema := skydb.RecordSchema{}
	for key, schema := range recordSchema {
		remoteSchema, ok := remoteRecordSchema[key]
		if !ok {
			updatingSchema[key] = schema
		} else if isConflict(remoteSchema, schema) {
			return fmt.Errorf("conflicting schema %s => %s", remoteSchema, schema)
		}

		// same data type, do nothing
	}

	if len(updatingSchema) > 0 {
		stmt := db.addColumnStmt(recordType, updatingSchema)

		log.WithField("stmt", stmt).Debugln("Adding columns to table")
		if _, err := db.Db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to alter table: %s", err)
		}
	}
	delete(db.c.RecordSchema, recordType)
	return nil
}

func (db *database) createTable(recordType string) (err error) {
	tablename := db.tableName(recordType)

	stmt := createTableStmt(tablename)
	log.WithField("stmt", stmt).Debugln("Creating table")
	_, err = db.Db.Exec(stmt)
	if err != nil {
		return err
	}

	const CreateTriggerStmtFmt = `CREATE TRIGGER trigger_notify_record_change
	AFTER INSERT OR UPDATE OR DELETE ON %s FOR EACH ROW
	EXECUTE PROCEDURE public.notify_record_change();
`
	stmt = fmt.Sprintf(CreateTriggerStmtFmt, tablename)
	log.WithField("stmt", stmt).Debugln("Creating trigger")
	_, err = db.Db.Exec(stmt)

	return err
}

func isConflict(from, to skydb.FieldType) bool {
	if from.Type == to.Type {
		return false
	}

	// currently integer can only be created by sequence,
	// so there are no conflicts
	if from.Type == skydb.TypeInteger && to.Type == skydb.TypeSequence {
		return false
	}

	// for manual assignment of sequence
	if from.Type == skydb.TypeInteger && to.Type == skydb.TypeNumber {
		return false
	}

	return true
}

func createTableStmt(tableName string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
	_id text,
	_database_id text,
	_owner_id text,
	_access jsonb,
	_created_at timestamp without time zone NOT NULL,
	_created_by text,
	_updated_at timestamp without time zone NOT NULL,
	_updated_by text,
	PRIMARY KEY(_id, _database_id, _owner_id),
	UNIQUE (_id)
);
`, tableName)
}

// ALTER TABLE app__.note add collection text;
// ALTER TABLE app__.note
// ADD CONSTRAINT fk_note_collection_collection
// FOREIGN KEY (collection)
// REFERENCES app__.collection(_id);
func (db *database) addColumnStmt(recordType string, recordSchema skydb.RecordSchema) string {
	buf := bytes.Buffer{}
	buf.Write([]byte("ALTER TABLE "))
	buf.WriteString(db.tableName(recordType))
	buf.WriteByte(' ')
	for column, schema := range recordSchema {
		buf.Write([]byte("ADD "))
		buf.WriteString(pq.QuoteIdentifier(column))
		buf.WriteByte(' ')
		buf.WriteString(pqDataType(schema.Type))
		buf.WriteByte(',')
		switch schema.Type {
		case skydb.TypeAsset:
			db.writeForeignKeyConstraint(&buf, column, "_asset", "id")
		case skydb.TypeReference:
			db.writeForeignKeyConstraint(&buf, column, schema.ReferenceType, "_id")
		}
	}

	// remote the last ','
	buf.Truncate(buf.Len() - 1)

	return buf.String()
}

func (db *database) writeForeignKeyConstraint(buf *bytes.Buffer, localCol, referent, remoteCol string) {
	buf.Write([]byte(`ADD CONSTRAINT `))
	buf.WriteString(pq.QuoteIdentifier(fmt.Sprintf(`fk_%s_%s_%s`, localCol, referent, remoteCol)))
	buf.Write([]byte(` FOREIGN KEY (`))
	buf.WriteString(pq.QuoteIdentifier(localCol))
	buf.Write([]byte(`) REFERENCES `))
	buf.WriteString(db.tableName(referent))
	buf.Write([]byte(` (`))
	buf.WriteString(pq.QuoteIdentifier(remoteCol))
	buf.Write([]byte(`),`))
}

func pqDataType(dataType skydb.DataType) string {
	switch dataType {
	default:
		panic(fmt.Sprintf("Unsupported dataType = %s", dataType))
	case skydb.TypeString, skydb.TypeAsset, skydb.TypeReference:
		return TypeString
	case skydb.TypeNumber:
		return TypeNumber
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
		sql = fmt.Sprintf("NOT (%s)", sql)
	}
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
