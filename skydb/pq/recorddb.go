package pq

import (
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

// GetByIDs using SQL IN cause
func (db *database) GetByIDs(ids []skydb.RecordID) (*skydb.Rows, error) {
	id := ids[0]
	log.Debugf("GetByIDs Type: %s", id.Type)
	typemap, err := db.remoteColumnTypes(id.Type)
	if err != nil {
		return nil, err
	}

	if len(typemap) == 0 { // record type has not been created
		return nil, skydb.ErrRecordNotFound
	}

	idStrs := []interface{}{}
	for _, recordID := range ids {
		if recordID.Key != "" {
			idStrs = append(idStrs, recordID.Key)
		}
	}
	inCause, inArgs := literalToSQLOperand(idStrs)

	query := db.selectQuery(id.Type, typemap).
		Where("_id IN "+inCause, inArgs...)

	sql, args, err := query.ToSql()
	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
		"err":  err,
	}).Infoln("Getting records by ID")

	rows, err := db.Db.Queryx(sql, args...)
	if err != nil {
		log.Debugf("Getting records by ID failed %v", err)
		return nil, err
	}
	return newRows(id.Type, typemap, rows, err)
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

	typemap, err = updateTypemapForQuery(query, typemap)
	if err != nil {
		return nil, err
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

func updateTypemapForQuery(query *skydb.Query, typemap skydb.RecordSchema) (skydb.RecordSchema, error) {
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
	return typemap, nil
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
