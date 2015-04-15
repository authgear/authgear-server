package pq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/oursky/ourd/oddb"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var underscoreRe = regexp.MustCompile(`[.:]`)

func toLowerAndUnderscore(s string) string {
	return underscoreRe.ReplaceAllLiteralString(strings.ToLower(s), "_")
}

func isUniqueViolated(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return true
	}

	return false
}

type userInfo struct {
	ID             string        `db:"id"`
	Email          string        `db:"email"`
	HashedPassword []byte        `db:"password"`
	Auth           authInfoValue `db:"auth"`
}

// authInfoValue implements sql.Valuer and sql.Scanner s.t.
// oddb.AuthInfo can be saved into and recovered from postgresql
type authInfoValue oddb.AuthInfo

func (auth authInfoValue) Value() (driver.Value, error) {
	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(auth); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (auth *authInfoValue) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		fmt.Errorf("oddb: unsupported Scan pair: %T -> %T", value, auth)
	}

	return json.Unmarshal(b, auth)
}

type conn struct {
	Db      *sqlx.DB
	appName string
}

func (c *conn) CreateUser(userinfo *oddb.UserInfo) error {
	const CreateUserTableFmt = `
CREATE TABLE IF NOT EXISTS %v._user (
	id varchar(255) PRIMARY KEY,
	email varchar(255),
	password varchar(255),
	auth json
);
`

	_, err := c.Db.Exec(fmt.Sprintf(CreateUserTableFmt, c.schemaName()))
	if err != nil {
		panic(err)
	}

	sql, args, err := psql.Insert(c.tableName("_user")).
		Columns("id", "email", "password", "auth").
		Values(userinfo.ID, userinfo.Email, userinfo.HashedPassword, authInfoValue(userinfo.Auth)).
		ToSql()
	if err != nil {
		panic(err)
	}

	_, err = c.Db.Exec(sql, args...)
	if isUniqueViolated(err) {
		return oddb.ErrUserDuplicated
	}

	return err
}

func (c *conn) GetUser(id string, userinfo *oddb.UserInfo) error {
	selectSql, args, err := psql.Select("email", "password", "auth").
		From(c.tableName("_user")).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		panic(err)
	}

	email, password, auth := "", []byte{}, authInfoValue{}
	err = c.Db.QueryRow(selectSql, args...).Scan(
		&email,
		&password,
		&auth,
	)
	if err == sql.ErrNoRows {
		return oddb.ErrUserNotFound
	}

	userinfo.ID = id
	userinfo.Email = email
	userinfo.HashedPassword = password
	userinfo.Auth = oddb.AuthInfo(auth)

	return err
}

func (c *conn) UpdateUser(userinfo *oddb.UserInfo) error {
	updateSql, args, err := psql.Update(c.tableName("_user")).
		Set("email", userinfo.Email).
		Set("password", userinfo.HashedPassword).
		Set("auth", authInfoValue(userinfo.Auth)).
		Where("id = ?", userinfo.ID).
		ToSql()
	if err != nil {
		panic(err)
	}

	result, err := c.Db.Exec(updateSql, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return oddb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) DeleteUser(id string) error {
	query, args, err := psql.Delete(c.tableName("_user")).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		panic(err)
	}

	result, err := c.Db.Exec(query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return oddb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows deleted, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) GetDevice(id string, device *oddb.Device) error { return nil }
func (c *conn) SaveDevice(device *oddb.Device) error           { return nil }
func (c *conn) DeleteDevice(id string) error                   { return nil }

func (c *conn) PublicDB() oddb.Database {
	return &database{
		Db: c.Db,
		c:  c,
	}
}

func (c *conn) PrivateDB(userKey string) oddb.Database {
	return &database{
		Db:     c.Db,
		c:      c,
		userID: userKey,
	}
}

func (c *conn) AddDBRecordHook(hook oddb.DBHookFunc) {}
func (c *conn) Close() error                         { return nil }

func (c *conn) schemaName() string {
	return "app_" + toLowerAndUnderscore(c.appName)
}

func (c *conn) tableName(table string) string {
	return c.schemaName() + "." + table
}

// Different data types in Postgres
// NOTE(limouren): varchar is missing because text can replace them,
// see the docs here: http://www.postgresql.org/docs/9.4/static/datatype-character.html
const (
	TypeString    = "text"
	TypeTimestamp = "timestamp without time zone"
	TypeNumber    = "double precision"
)

type database struct {
	Db     *sqlx.DB
	c      *conn
	userID string
}

func (db *database) Conn() oddb.Conn { return db.c }
func (db *database) ID() string      { return "" }

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

type rowsIter struct {
	recordtype string
	typemap    map[string]string
	rows       *sqlx.Rows
}

func (rowsi rowsIter) Close() error {
	return rowsi.rows.Close()
}

func (rowsi rowsIter) Next(record *oddb.Record) error {
	if rowsi.rows.Next() {
		columns, err := rowsi.rows.Columns()
		if err != nil {
			return err
		}

		values := make([]interface{}, 0, len(columns))
		for _, column := range columns {
			dataType, ok := rowsi.typemap[column]
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

		if err := rowsi.rows.Scan(values...); err != nil {
			return err
		}

		record.Type = rowsi.recordtype
		record.Data = map[string]interface{}{}

		for i, column := range columns {
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
	} else if rowsi.rows.Err() != nil {
		return rowsi.rows.Err()
	} else {
		return io.EOF
	}
}

func newRows(recordtype string, typemap map[string]string, rows *sqlx.Rows, err error) (*oddb.Rows, error) {
	if err != nil {
		return nil, err
	}

	return oddb.NewRows(rowsIter{recordtype, typemap, rows}), nil
}

func (db *database) GetMatchingSubscription(record *oddb.Record) []oddb.Subscription {
	return []oddb.Subscription{}
}
func (db *database) GetSubscription(key string, subscription *oddb.Subscription) error { return nil }
func (db *database) SaveSubscription(subscription *oddb.Subscription) error            { return nil }
func (db *database) DeleteSubscription(key string) error                               { return nil }

// schemaName is a convenient method to access parent conn's schemaName
func (db *database) schemaName() string {
	return db.c.schemaName()
}

// tableName is a convenient method to access parent conn's tableName
func (db *database) tableName(table string) string {
	return db.c.tableName(table)
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

// Open returns a new connection to postgresql implementation
func Open(appName, connString string) (oddb.Conn, error) {
	const CreateSchemaFmt = `CREATE SCHEMA IF NOT EXISTS %v`

	db, err := sqlx.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	stmt := fmt.Sprintf(CreateSchemaFmt, toLowerAndUnderscore("app_"+appName))
	if _, err := db.Exec(stmt); err != nil {
		return nil, err
	}

	return &conn{
		Db:      db,
		appName: appName,
	}, nil
}

func init() {
	oddb.Register("pq", oddb.DriverFunc(Open))
}

var (
	_ oddb.Conn     = &conn{}
	_ oddb.Database = &database{}

	_ driver.Valuer = authInfoValue{}
)
