package pq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"

	"github.com/oursky/ourd/oddb"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var underscoreRe = regexp.MustCompile(`[.:]`)

var initDBOnce sync.Once

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
	option  string
}

func (c *conn) CreateUser(userinfo *oddb.UserInfo) error {
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

func (c *conn) GetDevice(id string, device *oddb.Device) error {
	err := psql.Select("type", "token", "user_id").
		From(c.tableName("_device")).
		Where("id = ?", id).
		RunWith(c.Db.DB).
		QueryRow().
		Scan(&device.Type, &device.Token, &device.UserInfoID)

	if err == sql.ErrNoRows {
		return oddb.ErrDeviceNotFound
	} else if err != nil {
		return err
	}

	device.ID = id

	return nil
}

func (c *conn) SaveDevice(device *oddb.Device) error {
	if device.ID == "" || device.Token == "" || device.Type == "" || device.UserInfoID == "" {
		return errors.New("invalid device: empty id or token or type or user id")
	}

	pkData := map[string]interface{}{"id": device.ID}
	data := map[string]interface{}{
		"type":    device.Type,
		"token":   device.Token,
		"user_id": device.UserInfoID,
	}

	sql, args := upsertQuery(c.tableName("_device"), pkData, data)
	_, err := c.Db.Exec(sql, args...)

	return err
}

func (c *conn) DeleteDevice(id string) error {
	result, err := psql.Delete(c.tableName("_device")).
		Where("id = ?", id).
		RunWith(c.Db.DB).
		Exec()

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return oddb.ErrDeviceNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

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

func (c *conn) Close() error { return nil }

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
	db, err := sqlx.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %s", err)
	}

	// TODO: it might be desirable to init DB in start-up time.
	initDBOnce.Do(func() { mustInitDB(db) })

	if err := initAppDB(db, appName); err != nil {
		return nil, err
	}

	return &conn{
		Db:      db,
		appName: appName,
		option:  connString,
	}, nil
}

// mustInitDB initialize database objects shared across all schemata.
func mustInitDB(db *sqlx.DB) {
	const CreatePendingNotificationTableStmt = `CREATE TABLE IF NOT EXISTS pending_notification (
	id SERIAL NOT NULL PRIMARY KEY,
	op text NOT NULL,
	appname text NOT NULL,
	recordtype text NOT NULL,
	record jsonb NOT NULL
);
`
	const CreateNotificationTriggerFuncStmt = `CREATE OR REPLACE FUNCTION public.notify_record_change() RETURNS TRIGGER AS $$
	DECLARE
		affected_record RECORD;
		inserted_id integer;
	BEGIN
		IF (TG_OP = 'DELETE') THEN
			affected_record := OLD;
		ELSE
			affected_record := NEW;
		END IF;
		INSERT INTO pending_notification (op, appname, recordtype, record)
			VALUES (TG_OP, TG_TABLE_SCHEMA, TG_TABLE_NAME, row_to_json(affected_record)::jsonb)
			RETURNING id INTO inserted_id;
		PERFORM pg_notify('record_change', inserted_id::TEXT);
		RETURN affected_record;
	END;
$$ LANGUAGE plpgsql;
`

	db.MustExec(CreatePendingNotificationTableStmt)
	db.MustExec(CreateNotificationTriggerFuncStmt)
}

func initAppDB(db *sqlx.DB, appName string) error {
	const CreateSchemaFmt = `CREATE SCHEMA IF NOT EXISTS %v;`
	const CreateUserTableFmt = `
CREATE TABLE IF NOT EXISTS %v._user (
	id varchar(255) PRIMARY KEY,
	email varchar(255),
	password varchar(255),
	auth json
);
`
	const CreateDeviceTableFmt = `
CREATE TABLE IF NOT EXISTS %[1]v._device (
	id text PRIMARY KEY,
	user_id text REFERENCES %[1]v._user (id),
	type text NOT NULL,
	token text NOT NULL,
	UNIQUE (user_id, type, token)
);
`
	const CreateSubscriptionTableFmt = `
CREATE TABLE IF NOT EXISTS %[1]v._subscription (
	id text NOT NULL,
	user_id text NOT NULL,
	device_id text REFERENCES %[1]v._device (id) NOT NULL,
	type text NOT NULL,
	notification_info jsonb,
	query jsonb,
	PRIMARY KEY(user_id, device_id, id)
);
`

	schemaName := "app_" + toLowerAndUnderscore(appName)

	createSchemaStmt := fmt.Sprintf(CreateSchemaFmt, schemaName)
	if _, err := db.Exec(createSchemaStmt); err != nil {
		return fmt.Errorf("failed to create schema: %s", err)
	}

	createUserTableStmt := fmt.Sprintf(CreateUserTableFmt, schemaName)
	if _, err := db.Exec(createUserTableStmt); err != nil {
		return fmt.Errorf("failed to create user table: %s", err)
	}

	createDeviceTableStmt := fmt.Sprintf(CreateDeviceTableFmt, schemaName)
	if _, err := db.Exec(createDeviceTableStmt); err != nil {
		return fmt.Errorf("failed to create device table: %s", err)
	}

	createSubscriptionTableSmt := fmt.Sprintf(CreateSubscriptionTableFmt, schemaName)
	if _, err := db.Exec(createSubscriptionTableSmt); err != nil {
		return fmt.Errorf("failed to create subscription table: %s", err)
	}

	return nil
}

func init() {
	oddb.Register("pq", oddb.DriverFunc(Open))
}

var (
	_ oddb.Conn     = &conn{}
	_ oddb.Database = &database{}

	_ driver.Valuer = authInfoValue{}
)
