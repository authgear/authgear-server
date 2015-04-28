package pq

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"

	"github.com/oursky/ourd/oddb"
)

// This file implements Record related operations of the
// oddb/pq implementation.

// Different data types that can be saved in and loaded from postgreSQL
// NOTE(limouren): varchar is missing because text can replace them,
// see the docs here: http://www.postgresql.org/docs/9.4/static/datatype-character.html
const (
	TypeString    = "text"
	TypeNumber    = "double precision"
	TypeBoolean   = "boolean"
	TypeJSON      = "jsonb"
	TypeTimestamp = "timestamp without time zone"
)

func nilOrEmpty(s string) interface{} {
	if s == "" {
		return nil
	}

	return s
}

func (db *database) Get(id oddb.RecordID, record *oddb.Record) error {
	typemap, err := db.remoteColumnTypes(id.Type)
	if err != nil {
		return err
	}

	if len(typemap) == 0 { // record type has not been created
		return oddb.ErrRecordNotFound
	}

	sql, args, err := db.selectQuery(id.Type, typemap).ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debugln("Getting record")

	row := db.Db.QueryRowx(sql, args...)
	if err := newRecordScanner(id.Type, typemap, row).Scan(record); err != nil {
		return err
	}
	return nil
}

// Save attempts to do a upsert
func (db *database) Save(record *oddb.Record) error {
	if record.ID.Key == "" {
		return errors.New("db.save: got empty record id")
	}
	if record.ID.Type == "" {
		return fmt.Errorf("db.save %s: got empty record type", record.ID.Key)
	}

	tablename := db.tableName(record.ID.Type)

	data := map[string]interface{}{}
	data["_id"] = record.ID.Key
	data["_user_id"] = db.userID
	for key, value := range record.Data {
		data[`"`+key+`"`] = value
	}
	insert := psql.Insert(tablename).SetMap(sq.Eq(data))

	sql, args, err := insert.ToSql()
	if err != nil {
		panic(err)
	}

	logAction := "insert"
	_, err = db.Db.Exec(sql, args...)

	if isUniqueViolated(err) {
		update := psql.Update(tablename).Where("_id = ?", record.ID.Key).SetMap(sq.Eq(data))

		sql, args, err = update.ToSql()
		if err != nil {
			panic(err)
		}

		logAction = "update"
		_, err = db.Db.Exec(sql, args...)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Debugf("Failed to %s record", logAction)
	}

	record.UserID = db.userID

	return err
}

func (db *database) Delete(id oddb.RecordID) error {
	sql, args, err := psql.Delete(db.tableName("note")).
		Where("_id = ? AND _user_id = ?", id.Key, db.userID).
		ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debug("Executing SQL")

	result, err := db.Db.Exec(sql, args...)
	if isUndefinedTable(err) {
		return oddb.ErrRecordNotFound
	} else if err != nil {
		log.WithFields(log.Fields{
			"id":   id,
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Errorf("Failed to execute delete record statement")
		return fmt.Errorf("delete %s: failed to delete record", id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.WithFields(log.Fields{
			"id":   id,
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Errorln("Failed to fetch rowsAffected")
		return fmt.Errorf("delete %s: failed to retrieve deletion status", id)
	}

	if rowsAffected == 0 {
		return oddb.ErrRecordNotFound
	} else if rowsAffected > 1 {
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

func (db *database) Query(query *oddb.Query) (*oddb.Rows, error) {
	if query.Type == "" {
		return nil, errors.New("got empty query type")
	}

	typemap, err := db.remoteColumnTypes(query.Type)
	if err != nil {
		return nil, err
	}

	if len(typemap) == 0 { // record type has not been created
		return oddb.EmptyRows, nil
	}

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
	typemap    oddb.RecordSchema
	cs         columnsScanner
	columns    []string
	err        error
}

func newRecordScanner(recordType string, typemap oddb.RecordSchema, cs columnsScanner) *recordScanner {
	columns, err := cs.Columns()
	return &recordScanner{recordType, typemap, cs, columns, err}
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
			return fmt.Errorf("received unknown data type = %v for column = %s", dataType, column)
		case oddb.TypeNumber:
			var number sql.NullFloat64
			values = append(values, &number)
		case oddb.TypeString:
			var str sql.NullString
			values = append(values, &str)
		case oddb.TypeDateTime:
			var ts pq.NullTime
			values = append(values, &ts)
		case oddb.TypeBoolean:
			var boolean sql.NullBool
			values = append(values, &boolean)
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
		case *sql.NullBool:
			if svalue.Valid {
				record.Set(column, svalue.Bool)
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

func (rowsi rowsIter) Next(record *oddb.Record) error {
	if rowsi.rows.Next() {
		return rowsi.rs.Scan(record)
	} else if rowsi.rows.Err() != nil {
		return rowsi.rows.Err()
	} else {
		return io.EOF
	}
}

func newRows(recordType string, typemap oddb.RecordSchema, rows *sqlx.Rows, err error) (*oddb.Rows, error) {
	if err != nil {
		return nil, err
	}
	rs := newRecordScanner(recordType, typemap, rows)
	return oddb.NewRows(rowsIter{rows, rs}), nil
}

func (db *database) selectQuery(recordType string, typemap oddb.RecordSchema) sq.SelectBuilder {
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

func (db *database) remoteColumnTypes(recordType string) (oddb.RecordSchema, error) {
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

	typemap := oddb.RecordSchema{}

	var columnName, pqType string
	for rows.Next() {
		if err := rows.Scan(&columnName, &pqType); err != nil {
			return nil, err
		}

		var dataType oddb.DataType
		switch pqType {
		default:
			return nil, fmt.Errorf("received unknown data type = %s for column = %s", pqType, columnName)
		case TypeString:
			dataType = oddb.TypeString
		case TypeNumber:
			dataType = oddb.TypeNumber
		case TypeTimestamp:
			dataType = oddb.TypeDateTime
		}

		typemap[columnName] = dataType
	}

	return typemap, nil
}

func (db *database) Extend(recordType string, schema oddb.RecordSchema) error {
	remoteschema, err := db.remoteColumnTypes(recordType)
	if err != nil {
		return err
	}

	if len(remoteschema) == 0 {
		if err := db.createTable(recordType, schema); err != nil {
			return fmt.Errorf("failed to create table: %s", err)
		}
	} else {
		updatingSchema := oddb.RecordSchema{}
		for key, dataType := range schema {
			remoteDataType, ok := remoteschema[key]
			if !ok {
				updatingSchema[key] = dataType
			} else if remoteDataType != dataType {
				return fmt.Errorf("conflicting dataType %s => %s", remoteDataType, dataType)
			}

			// same data type, do nothing
		}

		if len(updatingSchema) > 0 {
			stmt := addColumnStmt(db.tableName(recordType), updatingSchema)

			log.WithField("stmt", stmt).Debugln("Adding columns to table")
			if _, err := db.Db.Exec(stmt); err != nil {
				return fmt.Errorf("failed to alter table: %s", err)
			}
		}
	}

	return nil
}

func (db *database) createTable(recordType string, schema oddb.RecordSchema) (err error) {
	tablename := db.tableName(recordType)

	stmt := createTableStmt(tablename, schema)
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

func createTableStmt(tableName string, schema oddb.RecordSchema) string {
	buf := bytes.Buffer{}
	buf.Write([]byte("CREATE TABLE "))
	buf.WriteString(tableName)
	buf.Write([]byte("(_id text, _user_id text,"))

	for recordType, dataType := range schema {
		buf.WriteByte('"')
		buf.WriteString(recordType)
		buf.WriteByte('"')
		buf.WriteByte(' ')
		buf.WriteString(pqDataType(dataType))
		buf.WriteByte(',')
	}

	buf.Write([]byte("PRIMARY KEY(_id, _user_id));"))

	return buf.String()
}

func addColumnStmt(tableName string, schema oddb.RecordSchema) string {
	buf := bytes.Buffer{}
	buf.Write([]byte("ALTER TABLE "))
	buf.WriteString(tableName)
	buf.WriteByte(' ')
	for recordType, dataType := range schema {
		buf.Write([]byte("ADD "))
		buf.WriteByte('"')
		buf.WriteString(recordType)
		buf.WriteByte('"')
		buf.WriteByte(' ')
		buf.WriteString(pqDataType(dataType))
		buf.WriteByte(',')
	}

	// remote the last ','
	buf.Truncate(buf.Len() - 1)

	return buf.String()
}

func pqDataType(dataType oddb.DataType) string {
	switch dataType {
	default:
		panic(fmt.Sprintf("Unsupported dataType = %s", dataType))
	case oddb.TypeString:
		return TypeString
	case oddb.TypeNumber:
		return TypeNumber
	case oddb.TypeDateTime:
		return TypeTimestamp
	case oddb.TypeBoolean:
		return TypeBoolean
	}
}
