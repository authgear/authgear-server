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
	const CreateTableFmt = `
CREATE TABLE IF NOT EXISTS %v.note (
	_id varchar(255),
	_user_id varchar(255),
	content text,
	createdDateTime timestamp,
	lastModified timestamp,
	noteOrder double precision,
	PRIMARY KEY(_id, _user_id)
);
`

	if record.Type != "note" {
		return errors.New("only record type 'note' is supported for now")
	}

	createTableStmt := fmt.Sprintf(CreateTableFmt, db.schemaName())
	if _, err := db.Db.Exec(createTableStmt); err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	tablename := db.tableName("note")

	data := map[string]interface{}{}
	data["_id"] = record.Key
	data["_user_id"] = db.userID
	for key, value := range record.Data {
		data[key] = value
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
	if query.Type != "note" {
		return nil, errors.New("only record type 'note' is supported for now")
	}

	q := psql.Select("_id", "content", "noteorder").
		From(db.tableName("note")).
		Where("_user_id = ?", db.userID)
	for _, sort := range query.Sorts {
		switch sort.Order {
		default:
			return nil, fmt.Errorf("unknown sort order = %v", sort.Order)
		case oddb.Asc:
			q = q.OrderBy(sort.KeyPath + " ASC")
		case oddb.Desc:
			q = q.OrderBy(sort.KeyPath + " DESC")
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
	return newRows(rows, err)
}

type rowsIter struct {
	rows *sqlx.Rows
}

func (rowsi rowsIter) Close() error {
	return rowsi.rows.Close()
}

func (rowsi rowsIter) Next(record *oddb.Record) error {
	if rowsi.rows.Next() {
		var (
			id, content sql.NullString
			noteOrder   sql.NullFloat64
		)

		err := rowsi.rows.Scan(&id, &content, &noteOrder)
		if err != nil {
			return err
		}

		if id.String == "" {
			return errors.New("got empty compulsory field '_id'")
		}

		record.Type = "note"
		record.Key = id.String
		record.Data = map[string]interface{}{}
		if content.Valid {
			record.Data["content"] = content.String
		}
		if noteOrder.Valid {
			record.Data["noteOrder"] = noteOrder.Float64
		}

		return nil
	} else if rowsi.rows.Err() != nil {
		return rowsi.rows.Err()
	} else {
		return io.EOF
	}
}

func newRows(rows *sqlx.Rows, err error) (*oddb.Rows, error) {
	if err != nil {
		return nil, err
	}

	return oddb.NewRows(rowsIter{rows}), nil
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
