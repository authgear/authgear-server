package pq

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"

	"github.com/oursky/ourd/oddb"
)

// This file implements Record related operations of the
// oddb/pq implementation.

// Different data types that can be saved in record
// NOTE(limouren): varchar is missing because text can replace them,
// see the docs here: http://www.postgresql.org/docs/9.4/static/datatype-character.html
const (
	TypeString    = "text"
	TypeTimestamp = "timestamp without time zone"
	TypeNumber    = "double precision"
)

func (db *database) Get(key string, record *oddb.Record) error {
	sql, args, err := sq.Select("*").From(db.tableName("note")).
		Where("_id = ? AND _user_id = ?", key, db.userID).
		ToSql()
	if err != nil {
		panic(err)
	}

	m := map[string]interface{}{}

	if err := db.Db.Get(&m, sql, args...); err != nil {
		return fmt.Errorf("get %v: %v", key, err)
	}

	delete(m, "_id")

	record.Key = key
	record.Type = "note"
	record.Data = m

	return nil
}

// Save attempts to do a upsert
func (db *database) Save(record *oddb.Record) error {
	if record.Key == "" {
		return errors.New("db.save: got empty record id")
	}
	if record.Type == "" {
		return fmt.Errorf("db.save %s: got empty record type", record.Key)
	}

	tablename := db.tableName(record.Type)
	typemap := deriveColumnTypes(record.Data)

	remotetypemap, err := db.remoteColumnTypes(record.Type)
	if err != nil {
		return err
	}

	if len(remotetypemap) == 0 {
		stmt := createTableStmt(tablename, typemap)

		if _, err := db.Db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	} else {
		// TODO(limouren): check diff and alter table here
	}

	data := map[string]interface{}{}
	data["_id"] = record.Key
	data["_user_id"] = db.userID
	for key, value := range record.Data {
		data[`"`+key+`"`] = value
	}
	insert := psql.Insert(tablename).SetMap(sq.Eq(data))

	sql, args, err := insert.ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debug("Inserting record")

	_, err = db.Db.Exec(sql, args...)

	if isUniqueViolated(err) {
		update := psql.Update(tablename).Where("_id = ?", record.Key).SetMap(sq.Eq(record.Data))

		sql, args, err = update.ToSql()
		if err != nil {
			panic(err)
		}

		log.WithFields(log.Fields{
			"sql":  sql,
			"args": args,
		}).Debug("Updating record")

		_, err = db.Db.Exec(sql, args...)
	}

	return err
}

func (db *database) Delete(key string) error {
	sql, args, err := psql.Delete(db.tableName("note")).
		Where("_id = ? AND _user_id = ?", key, db.userID).
		ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debug("Executing SQL")

	result, err := db.Db.Exec(sql, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return oddb.ErrRecordNotFound
	} else if rowsAffected > 1 {
		return fmt.Errorf("%v rows deleted, want 1", rowsAffected)
	}

	return err
}

func (db *database) Query(query *oddb.Query) (*oddb.Rows, error) {
	if query.Type == "" {
		return nil, errors.New("got empty query type")
	}

	typemap, err := db.remoteColumnTypes(query.Type)
	if err != nil {
		return nil, err
	}

	// remove _user_id, we won't need it in the result set
	delete(typemap, "_user_id")
	q := db.selectQuery(query.Type, typemap)
	for _, sort := range query.Sorts {
		switch sort.Order {
		default:
			return nil, fmt.Errorf("unknown sort order = %v", sort.Order)
		// NOTE(limouren): better to verify KeyPath as well
		case oddb.Asc:
			q = q.OrderBy(`"` + sort.KeyPath + `"` + " ASC")
		case oddb.Desc:
			q = q.OrderBy(`"` + sort.KeyPath + `"` + " DESC")
		}
	}

	sql, args, err := q.ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debugln("Querying record")

	rows, err := db.Db.Queryx(sql, args...)
	return newRows(query.Type, typemap, rows, err)
}

// columnsScanner wraps over sqlx.Rows and sqlx.Row to provide
// a consistent interface for column scanning.
type columnsScanner interface {
	Columns() ([]string, error)
	Scan(dest ...interface{}) error
}

type recordScanner struct {
	recordType string
	typemap    map[string]string
	cs         columnsScanner
	columns    []string
	err        error
}

func newRecordScanner(recordType string, typemap map[string]string, cs columnsScanner) recordScanner {
	columns, err := cs.Columns()
	return recordScanner{recordType, typemap, cs, columns, err}
}

func (rs *recordScanner) Scan(record *oddb.Record) error {
	if rs.err != nil {
		return rs.err
	}

	values := make([]interface{}, 0, len(rs.columns))
	for _, column := range rs.columns {
		dataType, ok := rs.typemap[column]
		if !ok {
			return fmt.Errorf("received unknown column = %s", column)
		}
		switch dataType {
		default:
			return fmt.Errorf("received unknown data type = %s for column = %s", dataType, column)
		case TypeNumber:
			var number sql.NullFloat64
			values = append(values, &number)
		case TypeString:
			var str sql.NullString
			values = append(values, &str)
		case TypeTimestamp:
			var ts pq.NullTime
			values = append(values, &ts)
		}
	}

	if err := rs.cs.Scan(values...); err != nil {
		rs.err = err
		return err
	}

	record.Type = rs.recordType
	record.Data = map[string]interface{}{}

	for i, column := range rs.columns {
		value := values[i]
		switch svalue := value.(type) {
		default:
			return fmt.Errorf("received unexpected scanned type = %T for column = %s", value, column)
		case *sql.NullFloat64:
			if svalue.Valid {
				record.Set(column, svalue.Float64)
			}
		case *sql.NullString:
			if svalue.Valid {
				record.Set(column, svalue.String)
			}
		case *pq.NullTime:
			if svalue.Valid {
				record.Set(column, svalue.Time)
			}
		}
	}

	return nil
}

type rowsIter struct {
	rows *sqlx.Rows
	rs   recordScanner
}

func (rowsi rowsIter) Close() error {
	return rowsi.rows.Close()
}

func (rowsi rowsIter) Next(record *oddb.Record) error {
	if rowsi.rows.Next() {
		return rowsi.rs.Scan(record)
	} else if rowsi.rows.Err() != nil {
		return rowsi.rows.Err()
	} else {
		return io.EOF
	}
}

func newRows(recordType string, typemap map[string]string, rows *sqlx.Rows, err error) (*oddb.Rows, error) {
	if err != nil {
		return nil, err
	}
	rs := newRecordScanner(recordType, typemap, rows)
	return oddb.NewRows(rowsIter{rows, rs}), nil
}

func (db *database) selectQuery(recordType string, typemap map[string]string) sq.SelectBuilder {
	columns := make([]string, 0, len(typemap))
	for column := range typemap {
		columns = append(columns, column)
	}

	q := psql.Select()
	for column := range typemap {
		q = q.Column(`"` + column + `"`)
	}

	q = q.From(db.tableName(recordType)).
		Where("_user_id = ?", db.userID)

	return q
}

func (db *database) remoteColumnTypes(recordType string) (map[string]string, error) {
	sql, args, err := psql.Select("column_name", "data_type").
		From("information_schema.columns").
		Where("table_schema = ? AND table_name = ?", db.schemaName(), recordType).ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debugln("Querying columns schema")

	rows, err := db.Db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	typemap := map[string]string{}

	var columnName, dataType string
	for rows.Next() {
		if err := rows.Scan(&columnName, &dataType); err != nil {
			return nil, err
		}

		switch dataType {
		default:
			return nil, fmt.Errorf("received unknown data type = %s for column = %s", dataType, columnName)
		case TypeString, TypeNumber, TypeTimestamp:
			// do nothing
		}

		typemap[columnName] = dataType
	}

	return typemap, nil
}

func deriveColumnTypes(m map[string]interface{}) map[string]string {
	typemap := map[string]string{}
	for key, value := range m {
		switch value.(type) {
		default:
			log.WithFields(log.Fields{
				"key":   key,
				"value": value,
			}).Panicf("got unrecgonized type = %T", value)
		case float64:
			typemap[key] = TypeNumber
		case string:
			typemap[key] = TypeString
		case time.Time:
			typemap[key] = TypeTimestamp
		}
	}

	return typemap
}

func createTableStmt(tableName string, typemap map[string]string) string {
	buf := bytes.Buffer{}
	buf.Write([]byte("CREATE TABLE "))
	buf.WriteString(tableName)
	buf.Write([]byte("(_id text, _user_id text,"))

	for column, dataType := range typemap {
		buf.WriteByte('"')
		buf.WriteString(column)
		buf.WriteByte('"')
		buf.WriteByte(' ')
		buf.WriteString(dataType)
		buf.WriteByte(',')
	}

	buf.Write([]byte("PRIMARY KEY(_id, _user_id));"))

	return buf.String()
}
