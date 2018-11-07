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
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record/pq/builder"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func (s *recordStore) UserRecordType() string {
	return "user"
}

func (s *recordStore) Get(id record.ID, r *record.Record) error {
	typemap, err := s.RemoteColumnTypes(id.Type)
	if err != nil {
		return err
	}

	if len(typemap) == 0 { // record type has not been created
		return record.ErrRecordNotFound
	}

	builder := s.selectQuery(s.sqlBuilder.Select(), id.Type, typemap).Where("_id = ?", id.Key)
	row := s.sqlExecutor.QueryRowWith(builder)
	if err := newRecordScanner(id.Type, typemap, row).Scan(r); err == sql.ErrNoRows {
		return record.ErrRecordNotFound
	} else if err != nil {
		return err
	}
	return nil
}

// GetByIDs using SQL IN cause
// GetByIDs only support one type of records at a time. If you want to query
// array of ids belongs to different type, you need to call this method multiple
// time.
func (s *recordStore) GetByIDs(ids []record.ID, accessControlOptions *record.AccessControlOptions) (*record.Rows, error) {
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

	s.logger.Debugf("GetByIDs Type: %s", recordType)
	typemap, err := s.RemoteColumnTypes(recordType)
	if err != nil {
		return nil, err
	}
	if len(typemap) == 0 {
		s.logger.Debugf("Record Type has not been created")
		return nil, record.ErrRecordNotFound
	}

	inCause, inArgs := builder.LiteralToSQLOperand(idStrs)
	query := s.selectQuery(s.sqlBuilder.Select(), recordType, typemap).
		Where(pq.QuoteIdentifier("_id")+" IN "+inCause, inArgs...)

	tableName := s.recordTableNameValue(recordType)
	if !accessControlOptions.BypassAccessControl {
		factory := builder.NewPredicateSqlizerFactory(s, s.sqlBuilder, recordType, tableName)
		aclSqlizer, err := factory.NewAccessControlSqlizer(accessControlOptions.ViewAsUser, record.ReadLevel)
		if err != nil {
			return nil, err
		}
		query = query.Where(aclSqlizer)
	}

	rows, err := s.sqlExecutor.QueryWith(query)
	if err != nil {
		s.logger.Debugf("Getting records by ID failed %v", err)
		return nil, err
	}
	return newRows(recordType, typemap, rows, err)
}

// Save attempts to do a upsert
func (s *recordStore) Save(r *record.Record) error {
	if r.ID.Key == "" {
		return errors.New("db.save: got empty record id")
	}
	if r.ID.Type == "" {
		return fmt.Errorf("db.save %s: got empty record type", r.ID.Key)
	}
	if r.OwnerID == "" {
		return fmt.Errorf("db.save %s: got empty OwnerID", r.ID.Key)
	}

	pkData := map[string]interface{}{
		"_id": r.ID.Key,
	}

	typemap, err := s.RemoteColumnTypes(r.ID.Type)
	if err != nil {
		return err
	}

	wrappers := map[string]func(string) string{}
	for column, fieldType := range typemap {
		if fieldType.Type == record.TypeGeometry {
			wrappers[column] = func(val string) string {
				return fmt.Sprintf("ST_GeomFromGeoJSON(%s)", val)
			}
		}
	}

	upsert := builder.UpsertQueryWithWrappers(s.recordFullTableName(r.ID.Type), pkData, convert(r), wrappers).
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

	if err := s.preSave(typemap, r); err != nil {
		return err
	}

	row := s.sqlExecutor.QueryRowWith(upsert)
	if err = newRecordScanner(r.ID.Type, typemap, row).Scan(r); err != nil {
		if db.IsUniqueViolated(err) {
			return skyerr.NewErrorf(
				skyerr.Duplicated,
				fmt.Sprintf("violate unique constraint"),
			)
		}

		if db.IsInvalidInputSyntax(err) {
			return skyerr.NewErrorf(
				skyerr.InvalidArgument,
				fmt.Sprintf("failed to save %s: %s", r.ID, err),
			)
		}
		return skyerr.MakeError(err)
	}

	return nil
}

func (s *recordStore) preSave(schema record.Schema, r *record.Record) error {
	const SetSequenceMaxValue = `SELECT setval($1, GREATEST(max(%v), $2)) FROM %v;`

	for key, value := range r.Data {
		// we are setting a sequence field
		if schema[key].Type == record.TypeSequence {
			selectSQL := fmt.Sprintf(SetSequenceMaxValue, pq.QuoteIdentifier(key), s.recordFullTableName(r.ID.Type))
			seqName := s.sqlBuilder.FullTableName(fmt.Sprintf(`%v_%v_seq`, r.ID.Type, key))
			if _, err := s.sqlExecutor.Exec(selectSQL, seqName, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func convert(r *record.Record) map[string]interface{} {
	m := map[string]interface{}{}
	for key, rawValue := range r.Data {
		switch value := rawValue.(type) {
		case []interface{}:
			m[key] = jsonSliceValue(value)
		case map[string]interface{}:
			m[key] = jsonMapValue(value)
		case *record.Asset:
			m[key] = assetValue(*value)
		case record.Reference:
			m[key] = referenceValue(value)
		case record.Location:
			m[key] = locationValue(value)
		case record.Geometry:
			m[key] = geometryValue(value)
		case record.Unknown:
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

func (s *recordStore) Delete(id record.ID) error {
	builder := s.sqlBuilder.Delete(s.recordFullTableName(id.Type)).
		Where("_id = ?", id.Key)

	result, err := s.sqlExecutor.ExecWith(builder)
	if db.IsUndefinedTable(err) {
		return record.ErrRecordNotFound
	} else if db.IsForeignKeyViolated(err) {
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
		return record.ErrRecordNotFound
	} else if rowsAffected > 1 {
		s.logger.WithFields(logrus.Fields{
			"id":           id,
			"rowsAffected": rowsAffected,
			"err":          err,
		}).Errorln("Unexpected rows deleted")
		return fmt.Errorf("delete %s: got %v rows deleted, want 1", id, rowsAffected)
	}

	return err
}

func (s *recordStore) applyQueryPredicate(q sq.SelectBuilder, factory builder.PredicateSqlizerFactory, query *record.Query, accessControlOptions *record.AccessControlOptions) (sq.SelectBuilder, error) {
	if p := query.Predicate; !p.IsEmpty() {
		sqlizer, err := factory.NewPredicateSqlizer(p)
		if err != nil {
			return q, err
		}
		q = q.Where(sqlizer)
		q = factory.AddJoinsToSelectBuilder(q)
	}

	if !accessControlOptions.BypassAccessControl {
		aclSqlizer, err := factory.NewAccessControlSqlizer(accessControlOptions.ViewAsUser, record.ReadLevel)
		if err != nil {
			return q, err
		}
		q = q.Where(aclSqlizer)
	}

	return q, nil
}

func (s *recordStore) Query(query *record.Query, accessControlOptions *record.AccessControlOptions) (*record.Rows, error) {
	if query.Type == "" {
		return nil, errors.New("got empty query type")
	}

	typemap, err := s.RemoteColumnTypes(query.Type)
	if err != nil {
		return nil, err
	}

	if len(typemap) == 0 { // record type has not been created
		return record.EmptyRows, nil
	}

	tableName := s.recordTableNameValue(query.Type)
	// tableName := query.Type

	q := s.sqlBuilder.Select()
	factory := builder.NewPredicateSqlizerFactory(s, s.sqlBuilder, query.Type, tableName)
	q, err = s.applyQueryPredicate(q, factory, query, accessControlOptions)
	if err != nil {
		return nil, err
	}

	for _, sort := range query.Sorts {
		orderBy, err := builder.SortOrderBySQL(tableName, sort)
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
	q = s.selectQuery(q, query.Type, typemap)

	rows, err := s.sqlExecutor.QueryWith(q)
	return newRows(query.Type, typemap, rows, err)
}

func (s *recordStore) QueryCount(query *record.Query, accessControlOptions *record.AccessControlOptions) (uint64, error) {
	if query.Type == "" {
		return 0, errors.New("got empty query type")
	}

	typemap, err := s.RemoteColumnTypes(query.Type)
	if err != nil || len(typemap) == 0 { // error or record type has not been created
		return 0, err
	}

	typemap = record.Schema{
		"_record_count": record.FieldType{
			Type: record.TypeNumber,
			Expression: record.Expression{
				Type: record.Function,
				Value: record.CountFunc{
					OverallRecords: false,
				},
			},
		},
	}

	tableName := s.recordTableNameValue(query.Type)
	q := s.selectQuery(s.sqlBuilder.Select(), query.Type, typemap)
	factory := builder.NewPredicateSqlizerFactory(s, s.sqlBuilder, query.Type, tableName)
	q, err = s.applyQueryPredicate(q, factory, query, accessControlOptions)
	if err != nil {
		return 0, err
	}

	rows, err := s.sqlExecutor.QueryWith(q)
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
	typemap     record.Schema
	cs          columnsScanner
	columns     []string
	err         error
	recordCount *uint64
}

func newRecordScanner(recordType string, typemap record.Schema, cs columnsScanner) *recordScanner {
	columns, err := cs.Columns()
	return &recordScanner{recordType, typemap, cs, columns, err, nil}
}

// nolint: gocyclo
func (rs *recordScanner) Scan(r *record.Record) error {
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
		case record.TypeNumber:
			var number sql.NullFloat64
			values = append(values, &number)
		case record.TypeString, record.TypeReference, record.TypeACL:
			var str sql.NullString
			values = append(values, &str)
		case record.TypeDateTime:
			var ts pq.NullTime
			values = append(values, &ts)
		case record.TypeBoolean:
			var boolean sql.NullBool
			values = append(values, &boolean)
		case record.TypeAsset:
			var asset nullAsset
			values = append(values, &asset)
		case record.TypeJSON:
			var j nullJSON
			values = append(values, &j)
		case record.TypeLocation:
			var l nullLocation
			values = append(values, &l)
		case record.TypeSequence:
			fallthrough
		case record.TypeInteger:
			var i sql.NullInt64
			values = append(values, &i)
		case record.TypeGeometry:
			var g nullGeometry
			values = append(values, &g)
		case record.TypeUnknown:
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

	r.ID.Type = rs.recordType
	r.Data = map[string]interface{}{}

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
				r.Set(column, svalue.Float64)
			}
		case *sql.NullString:
			if svalue.Valid {
				schema := rs.typemap[column]
				if schema.Type == record.TypeReference {
					r.Set(column, record.NewReference(schema.ReferenceType, svalue.String))
				} else if schema.Type == record.TypeACL {
					acl := record.ACL{}
					json.Unmarshal([]byte(svalue.String), &acl)
					r.Set(column, acl)
				} else {
					r.Set(column, svalue.String)
				}
			}
		case *pq.NullTime:
			if svalue.Valid {
				// it is to support direct deep-equal of value between
				// a empty record and a record materialized from the database
				if svalue.Time.IsZero() {
					r.Set(column, time.Time{})
				} else {
					r.Set(column, svalue.Time.In(time.UTC))
				}
			}
		case *sql.NullBool:
			if svalue.Valid {
				r.Set(column, svalue.Bool)
			}
		case *nullAsset:
			if svalue.Valid {
				r.Set(column, svalue.Asset)
			}
		case *nullJSON:
			if svalue.Valid {
				r.Set(column, svalue.JSON)
			}
		case *nullLocation:
			if svalue.Valid {
				r.Set(column, svalue.Location)
			}
		case *nullGeometry:
			if svalue.Valid {
				r.Set(column, svalue.Geometry)
			}
		case *nullUnknown:
			if svalue.Valid {
				val := record.Unknown{}
				if schema, ok := rs.typemap[column]; ok {
					val.UnderlyingType = schema.UnderlyingType
				}
				r.Set(column, val)
			}
		case *sql.NullInt64:
			if svalue.Valid {
				r.Set(column, svalue.Int64)
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

func (rowsi rowsIter) Next(record *record.Record) error {
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

func newRows(recordType string, typemap record.Schema, rows *sqlx.Rows, err error) (*record.Rows, error) {
	if err != nil {
		return nil, err
	}
	rs := newRecordScanner(recordType, typemap, rows)
	return record.NewRows(rowsIter{rows, rs}), nil
}

func columnSqlizersForSelect(alias string, typemap record.Schema) map[string]sq.Sqlizer {
	sqlizers := map[string]sq.Sqlizer{}
	for column, fieldType := range typemap {
		expr := fieldType.Expression
		if expr.IsEmpty() {
			expr = record.Expression{
				Type:  record.KeyPath,
				Value: column,
			}
		}

		sqlizer := builder.NewExpressionSqlizer(alias, fieldType, expr)
		if fieldType.Type == record.TypeGeometry {
			sqlizer, _ = builder.RequireCast(sqlizer)
		}
		sqlizers[column] = sqlizer
	}
	return sqlizers
}

func (s *recordStore) selectQuery(q sq.SelectBuilder, recordType string, typemap record.Schema) sq.SelectBuilder {
	for column, e := range columnSqlizersForSelect(s.recordTableNameValue(recordType), typemap) {
		sqlOperand, opArgs, _ := e.ToSql()
		q = q.Column(sqlOperand+" as "+pq.QuoteIdentifier(column), opArgs...)
	}

	q = q.From(s.recordFullTableName(recordType))

	return q
}

func updateTypemapForQuery(query *record.Query, typemap record.Schema) (record.Schema, error) {
	if query.DesiredKeys != nil {
		newtypemap, err := whitelistedRecordSchema(typemap, query.DesiredKeys)
		if err != nil {
			return nil, err
		}
		typemap = newtypemap
	}

	for key, value := range query.ComputedKeys {
		if value.Type == record.KeyPath {
			// recorddb does not support querying with computed keys
			continue
		}

		v := value // because value will be overwritten in the next loop
		typemap["_transient_"+key] = record.FieldType{
			Type:       record.TypeNumber,
			Expression: v,
		}
	}

	if query.GetCount {
		typemap["_record_count"] = record.FieldType{
			Type: record.TypeNumber,
			Expression: record.Expression{
				Type: record.Function,
				Value: record.CountFunc{
					OverallRecords: true,
				},
			},
		}
	}
	return typemap, nil
}

func whitelistedRecordSchema(schema record.Schema, whitelistKeys []string) (record.Schema, error) {
	wlSchema := record.Schema{}

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
