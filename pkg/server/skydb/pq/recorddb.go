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
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/pq/builder"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func (db *database) Get(id skydb.RecordID, record *skydb.Record) error {
	typemap, err := db.RemoteColumnTypes(id.Type)
	if err != nil {
		return err
	}

	if len(typemap) == 0 { // record type has not been created
		return skydb.ErrRecordNotFound
	}

	builder := db.selectQuery(psql.Select(), id.Type, typemap).Where("_id = ?", id.Key)
	row := db.c.QueryRowWith(builder)
	if err := newRecordScanner(id.Type, typemap, row).Scan(record); err == sql.ErrNoRows {
		return skydb.ErrRecordNotFound
	} else if err != nil {
		return err
	}
	return nil
}

// GetByIDs using SQL IN cause
// GetByIDs only support one type of records at a time. If you want to query
// array of ids belongs to different type, you need to call this method multiple
// time.
func (db *database) GetByIDs(ids []skydb.RecordID, accessControlOptions *skydb.AccessControlOptions) (*skydb.Rows, error) {
	if len(ids) == 0 {
		return nil, errors.New("db.GetByIDs received empty array")
	}
	idStrs := []interface{}{}
	recordType := ""
	for _, recordID := range ids {
		if recordID.Key != "" {
			idStrs = append(idStrs, recordID.Key)
		}
		if recordID.Type != "" && recordType == "" {
			recordType = recordID.Type
		}
	}

	log.Debugf("GetByIDs Type: %s", recordType)
	typemap, err := db.RemoteColumnTypes(recordType)
	if err != nil {
		return nil, err
	}
	if len(typemap) == 0 {
		log.Debugf("Record Type has not been created")
		return nil, skydb.ErrRecordNotFound
	}

	inCause, inArgs := builder.LiteralToSQLOperand(idStrs)
	query := db.selectQuery(psql.Select(), recordType, typemap).
		Where(pq.QuoteIdentifier("_id")+" IN "+inCause, inArgs...)

	if db.DatabaseType() == skydb.PublicDatabase && !accessControlOptions.BypassAccessControl {
		factory := builder.NewPredicateSqlizerFactory(db, recordType)
		aclSqlizer, err := factory.NewAccessControlSqlizer(accessControlOptions.ViewAsUser, skydb.ReadLevel)
		if err != nil {
			return nil, err
		}
		query = query.Where(aclSqlizer)
	}

	rows, err := db.c.QueryWith(query)
	if err != nil {
		log.Debugf("Getting records by ID failed %v", err)
		return nil, err
	}
	return newRows(recordType, typemap, rows, err)
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

	var pkData map[string]interface{}
	switch db.DatabaseType() {
	case skydb.UnionDatabase:
		return skydb.ErrDatabaseIsReadOnly
	case skydb.PublicDatabase:
		fallthrough
	case skydb.PrivateDatabase:
		pkData = map[string]interface{}{
			"_id":          record.ID.Key,
			"_database_id": db.userID,
		}
	}

	typemap, err := db.RemoteColumnTypes(record.ID.Type)
	if err != nil {
		return err
	}

	wrappers := map[string]func(string) string{}
	for column, fieldType := range typemap {
		if fieldType.Type == skydb.TypeGeometry {
			wrappers[column] = func(val string) string {
				return fmt.Sprintf("ST_GeomFromGeoJSON(%s)", val)
			}
		}
	}

	upsert := builder.UpsertQueryWithWrappers(db.TableName(record.ID.Type), pkData, convert(record), wrappers).
		IgnoreKeyOnUpdate("_owner_id").
		IgnoreKeyOnUpdate("_created_at").
		IgnoreKeyOnUpdate("_created_by")

	// record type is empty in the following statement because upsert
	// only concerns with one record type, and that specifying the
	// name of the record type here actually causes the SQL to find
	// the table, which is not found because aliasing.
	for column, sqlizer := range columnSqlizersForSelect("", typemap) {
		upsert = upsert.SelectColumn(column, sqlizer)
	}

	if err := db.preSave(typemap, record); err != nil {
		return err
	}

	row := db.c.QueryRowWith(upsert)
	if err = newRecordScanner(record.ID.Type, typemap, row).Scan(record); err != nil {
		if isUniqueViolated(err) {
			return skyerr.NewErrorf(
				skyerr.Duplicated,
				fmt.Sprintf("violate unique constraint"),
			)
		}

		if isInvalidInputSyntax(err) {
			return skyerr.NewErrorf(
				skyerr.InvalidArgument,
				fmt.Sprintf("failed to save %s: %s", record.ID, err),
			)
		}
		return skyerr.MakeError(err)
	}

	record.DatabaseID = db.userID
	return nil
}

func (db *database) preSave(schema skydb.RecordSchema, record *skydb.Record) error {
	const SetSequenceMaxValue = `SELECT setval($1, GREATEST(max(%v), $2)) FROM %v;`

	for key, value := range record.Data {
		// we are setting a sequence field
		if schema[key].Type == skydb.TypeSequence {
			selectSQL := fmt.Sprintf(SetSequenceMaxValue, pq.QuoteIdentifier(key), db.TableName(record.ID.Type))
			seqName := db.TableName(fmt.Sprintf(`%v_%v_seq`, record.ID.Type, key))
			if _, err := db.c.Exec(selectSQL, seqName, value); err != nil {
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
		case skydb.Location:
			m[key] = locationValue(value)
		case skydb.Geometry:
			m[key] = geometryValue(value)
		case skydb.Unknown:
			// Do not modify columns with unknown type because they are
			// managed by the developer.
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
	builder := psql.Delete(db.TableName(id.Type)).
		Where("_id = ?", id.Key)

	switch db.DatabaseType() {
	case skydb.UnionDatabase:
		return skydb.ErrDatabaseIsReadOnly
	case skydb.PublicDatabase:
		fallthrough
	case skydb.PrivateDatabase:
		builder = builder.Where("_database_id = ?", db.userID)
	}

	result, err := db.c.ExecWith(builder)
	if isUndefinedTable(err) {
		return skydb.ErrRecordNotFound
	} else if isForeignKeyViolated(err) {
		return skyerr.NewError(
			skyerr.ConstraintViolated,
			fmt.Sprintf("delete %s: failed to delete record because other records have reference to it", id),
		)
	} else if err != nil {
		return fmt.Errorf("delete %s: failed to delete record", id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete %s: failed to retrieve deletion status", id)
	}

	if rowsAffected == 0 {
		return skydb.ErrRecordNotFound
	} else if rowsAffected > 1 {
		log.WithFields(logrus.Fields{
			"id":           id,
			"rowsAffected": rowsAffected,
			"err":          err,
		}).Errorln("Unexpected rows deleted")
		return fmt.Errorf("delete %s: got %v rows deleted, want 1", id, rowsAffected)
	}

	return err
}

func (db *database) applyQueryPredicate(q sq.SelectBuilder, factory builder.PredicateSqlizerFactory, query *skydb.Query, accessControlOptions *skydb.AccessControlOptions) (sq.SelectBuilder, error) {
	if p := query.Predicate; !p.IsEmpty() {
		sqlizer, err := factory.NewPredicateSqlizer(p)
		if err != nil {
			return q, err
		}
		q = q.Where(sqlizer)
		q = factory.AddJoinsToSelectBuilder(q)
	}

	if db.DatabaseType() == skydb.PublicDatabase && !accessControlOptions.BypassAccessControl {
		aclSqlizer, err := factory.NewAccessControlSqlizer(accessControlOptions.ViewAsUser, skydb.ReadLevel)
		if err != nil {
			return q, err
		}
		q = q.Where(aclSqlizer)
	}

	return q, nil
}

func (db *database) Query(query *skydb.Query, accessControlOptions *skydb.AccessControlOptions) (*skydb.Rows, error) {
	if query.Type == "" {
		return nil, errors.New("got empty query type")
	}

	typemap, err := db.RemoteColumnTypes(query.Type)
	if err != nil {
		return nil, err
	}

	if len(typemap) == 0 { // record type has not been created
		return skydb.EmptyRows, nil
	}

	q := psql.Select()
	factory := builder.NewPredicateSqlizerFactory(db, query.Type)
	q, err = db.applyQueryPredicate(q, factory, query, accessControlOptions)
	if err != nil {
		return nil, err
	}

	for _, sort := range query.Sorts {
		orderBy, err := builder.SortOrderBySQL(query.Type, sort)
		if err != nil {
			return nil, err
		}
		q = q.OrderBy(orderBy)
	}

	if query.Limit != nil {
		q = q.Limit(*query.Limit)
	}

	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}

	// Select columns to return, this is the last step so that predicate
	// have a chance to change typemap, and the generated sql
	// depends on the alias name used in table joins.
	typemap, err = updateTypemapForQuery(query, typemap)
	if err != nil {
		return nil, err
	}
	typemap = factory.UpdateTypemap(typemap)
	q = db.selectQuery(q, query.Type, typemap)

	rows, err := db.c.QueryWith(q)
	return newRows(query.Type, typemap, rows, err)
}

func (db *database) QueryCount(query *skydb.Query, accessControlOptions *skydb.AccessControlOptions) (uint64, error) {
	if query.Type == "" {
		return 0, errors.New("got empty query type")
	}

	typemap, err := db.RemoteColumnTypes(query.Type)
	if err != nil || len(typemap) == 0 { // error or record type has not been created
		return 0, err
	}

	typemap = skydb.RecordSchema{
		"_record_count": skydb.FieldType{
			Type: skydb.TypeNumber,
			Expression: skydb.Expression{
				Type: skydb.Function,
				Value: skydb.CountFunc{
					OverallRecords: false,
				},
			},
		},
	}

	q := db.selectQuery(psql.Select(), query.Type, typemap)
	factory := builder.NewPredicateSqlizerFactory(db, query.Type)
	q, err = db.applyQueryPredicate(q, factory, query, accessControlOptions)
	if err != nil {
		return 0, err
	}

	rows, err := db.c.QueryWith(q)
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

// nolint: gocyclo
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
		case skydb.TypeSequence:
			fallthrough
		case skydb.TypeInteger:
			var i sql.NullInt64
			values = append(values, &i)
		case skydb.TypeGeometry:
			var g nullGeometry
			values = append(values, &g)
		case skydb.TypeUnknown:
			var u nullUnknown
			values = append(values, &u)
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
				record.Set(column, svalue.Location)
			}
		case *nullGeometry:
			if svalue.Valid {
				record.Set(column, svalue.Geometry)
			}
		case *nullUnknown:
			if svalue.Valid {
				val := skydb.Unknown{}
				if schema, ok := rs.typemap[column]; ok {
					val.UnderlyingType = schema.UnderlyingType
				}
				record.Set(column, val)
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

func columnSqlizersForSelect(recordType string, typemap skydb.RecordSchema) map[string]sq.Sqlizer {
	sqlizers := map[string]sq.Sqlizer{}
	for column, fieldType := range typemap {
		expr := fieldType.Expression
		if expr.IsEmpty() {
			expr = skydb.Expression{
				Type:  skydb.KeyPath,
				Value: column,
			}
		}

		sqlizer := builder.NewExpressionSqlizer(recordType, fieldType, expr)
		if fieldType.Type == skydb.TypeGeometry {
			sqlizer, _ = builder.RequireCast(sqlizer)
		}
		sqlizers[column] = sqlizer
	}
	return sqlizers
}

func (db *database) selectQuery(q sq.SelectBuilder, recordType string, typemap skydb.RecordSchema) sq.SelectBuilder {
	for column, e := range columnSqlizersForSelect(recordType, typemap) {
		sqlOperand, opArgs, _ := e.ToSql()
		q = q.Column(sqlOperand+" as "+pq.QuoteIdentifier(column), opArgs...)
	}

	q = q.From(db.TableName(recordType))

	switch db.DatabaseType() {
	case skydb.UnionDatabase:
		// no filter on `_database_id` column
	case skydb.PublicDatabase:
		fallthrough
	case skydb.PrivateDatabase:
		q = q.Where(fmt.Sprintf(`%s."_database_id" = ?`, pq.QuoteIdentifier(recordType)), db.userID)
	}
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

		v := value // because value will be overwritten in the next loop
		typemap["_transient_"+key] = skydb.FieldType{
			Type:       skydb.TypeNumber,
			Expression: v,
		}
	}

	if query.GetCount {
		typemap["_record_count"] = skydb.FieldType{
			Type: skydb.TypeNumber,
			Expression: skydb.Expression{
				Type: skydb.Function,
				Value: skydb.CountFunc{
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
