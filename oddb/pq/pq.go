package pq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
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

func isUndefinedTable(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42P01" {
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
